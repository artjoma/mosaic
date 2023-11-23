package api

import (
	"encoding/hex"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/pierrec/lz4/v4"
	"log/slog"
	"mosaic/engine"
	"net/http"
	"strconv"
)

var NotFoundErrMsg = []byte("Page Not Found")

type HttpApi struct {
	address    string
	port       uint16
	engine     *engine.Engine
	httpRouter *fiber.App
}

/*
Create new Http Api instance
*/
func NewHttpApi(address string, port uint16, engine *engine.Engine) *HttpApi {
	return &HttpApi{
		port:    port,
		address: address,
		engine:  engine,
	}
}

// SetupHttpServer block caller thread !
func (api *HttpApi) SetupHttpServer() {
	bindAddr := api.address + ":" + strconv.Itoa(int(api.port))
	slog.Info("Setup HTTP API", "address", bindAddr)
	router := fiber.New(
		fiber.Config{
			StreamRequestBody: true,
		})
	api.httpRouter = router

	router.Use(recover.New())
	v1 := router.Group("/v1")
	// curl -X POST -H "Content-Type: application/json" -d '{"host":"0.0.0.0:6751"}' http://0.0.0.0:25010/v1/shard/add
	v1.Post("/shard/add", api.addShardHandler)
	// curl -F file=@OIPD.pdf http://0.0.0.0:25010/v1/file/put
	v1.Post("/file/put", api.putFileHandler)
	// curl http://0.0.0.0:25010/v1/file/download/<fileId>
	v1.Get("/file/download/:fId", api.downloadFileHandler)
	// curl http://0.0.0.0:25010/v1/file/meta/:fileId
	v1.Get("/file/meta/:fId", api.getFileMetadata)
	// curl http://0.0.0.0:25010/v1/cluster/state
	v1.Get("/cluster/state", api.getClusterState)
	// Last middleware to match anything
	v1.Use(func(c *fiber.Ctx) error {
		return c.SendStatus(404) // => 404 "Not Found"
	})
	if err := router.Listen(bindAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Http API", "err", err.Error())
	}
}

func (api *HttpApi) addShardHandler(ctx *fiber.Ctx) error {
	req := &AddShardRequest{}
	if err := ctx.BodyParser(req); err != nil {
		return err
	}
	if req.Host == "" {
		return api.writeErrResponse(ctx, http.StatusBadRequest, errors.New("invalid request"))
	}
	cmd := &engine.AddShardCmd{
		Host: req.Host,
	}
	if err := cmd.Prepare(api.engine); err != nil {
		return api.writeErrResponse(ctx, http.StatusInternalServerError, err)
	}

	api.engine.ExecuteCmdAsync(cmd)

	return api.writeResponse(ctx, &AddShardResponse{
		ShardId: cmd.Id,
	})
}

func (api *HttpApi) putFileHandler(ctx *fiber.Ctx) error {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return api.writeErrResponse(ctx, http.StatusInternalServerError, err)
	}

	// Check if empty body or not
	if fileHeader.Size < 2 {
		return api.writeErrResponse(ctx, http.StatusBadRequest, errors.New("got content length < 2"))
	}
	f, err := fileHeader.Open()
	defer f.Close()

	cmd := &engine.PutFileCmd{}
	cmd.SetPrepare(f, fileHeader.Filename)

	if cmd.Prepare(api.engine) != nil {
		return api.writeErrResponse(ctx, http.StatusInternalServerError, err)
	}

	api.engine.ExecuteCmdAsync(cmd)
	return api.writeResponse(ctx, &PutFileResponse{
		cmd.FileMetadata,
	})
}

func (api *HttpApi) downloadFileHandler(ctx *fiber.Ctx) error {
	fIdParam := ctx.Params("fId")
	if fIdParam == "" {
		return api.writeErrResponse(ctx, http.StatusBadRequest, errors.New("invalid request"))
	}
	fId, err := hex.DecodeString(fIdParam)
	if err != nil {
		return api.writeErrResponse(ctx, http.StatusBadRequest, errors.New("invalid request"))
	}
	_, buff, err := api.engine.DownloadFile(fId)

	zr := lz4.NewReader(buff)
	if err := zr.Apply(lz4.ConcurrencyOption(4)); err != nil {
		return err
	}

	ctx.SendStream(zr)
	return err
}

// TODO For browsers use attachment+file name
func (api *HttpApi) downloadFileAttachHandler(ctx *fiber.Ctx) error {
	return nil
}

func (api *HttpApi) getFileMetadata(ctx *fiber.Ctx) error {
	fIdParam := ctx.Params("fId")
	if fIdParam == "" {
		return api.writeErrResponse(ctx, http.StatusBadRequest, errors.New("invalid request"))
	}
	fId, err := hex.DecodeString(fIdParam)
	if err != nil {
		return api.writeErrResponse(ctx, http.StatusBadRequest, errors.New("invalid request"))
	}
	fMeta, err := api.engine.GetFileMetadata(fId)
	if err != nil {
		return api.writeErrResponse(ctx, http.StatusInternalServerError, err)
	}

	return api.writeResponse(ctx, fMeta)
}

func (api *HttpApi) getClusterState(ctx *fiber.Ctx) error {
	state, err := api.engine.ClusterState()
	if err != nil {
		return api.writeErrResponse(ctx, http.StatusInternalServerError, err)
	}

	return api.writeResponse(ctx, state)
}

func (api *HttpApi) writeErrResponse(ctx *fiber.Ctx, httpCode int, err error) error {
	return ctx.Status(httpCode).JSON(NewErrResponse(err))
}

func (api *HttpApi) writeResponse(ctx *fiber.Ctx, model interface{}) error {
	return ctx.Status(200).JSON(model)
}
