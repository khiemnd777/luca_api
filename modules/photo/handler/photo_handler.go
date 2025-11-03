// scripts/create_module/templates/handler_http.go.tmpl
package handler

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/photo/config"
	"github.com/khiemnd777/andy_api/modules/photo/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type PhotoHandler struct {
	svc  *service.PhotoService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewPhotoHandler(svc *service.PhotoService, deps *module.ModuleDeps[config.ModuleConfig]) *PhotoHandler {
	return &PhotoHandler{
		svc:  svc,
		deps: deps,
	}
}

func (h *PhotoHandler) RegisterRoutes(router fiber.Router) {
	app.RouterPost(router, "/", h.UploadPhoto)
	app.RouterGet(router, "/", h.GetPhotos)

	app.RouterGet(router, "/file/:filename", h.GetPhotoFile)

	app.RouterPost(router, "/batch-get", h.BatchGet)
	app.RouterPost(router, "/update-folder", h.UpdateFolder)
	app.RouterPost(router, "/delete-batch", h.DeleteMany)

	app.RouterGet(router, "/:id", h.GetPhoto)
	app.RouterGet(router, "/name/:filename", h.GetPhotoByFileName)
	app.RouterDelete(router, "/:id", h.DeletePhoto)
}

func (h *PhotoHandler) UploadPhoto(c *fiber.Ctx) error {
	file, err := c.FormFile("photo")
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Missing photo file")
	}

	userId, _ := strconv.Atoi(c.FormValue("user_id"))

	var folderId *int
	folderIDStr := c.FormValue("folder_id")
	if folderIDStr != "" {
		if id, err := strconv.Atoi(folderIDStr); err == nil {
			folderId = &id
		}
	}

	var capturedAt any
	capturedAtStr := c.FormValue("capturedAt")
	if capturedAtStr != "" {
		if t, err := utils.ParseDate(capturedAtStr); err == nil {
			capturedAt = t
		}
	}

	meta := map[string]any{
		"device":     c.FormValue("device"),
		"os":         c.FormValue("os"),
		"lat":        utils.ParseFloat(c.FormValue("lat")),
		"lng":        utils.ParseFloat(c.FormValue("lng")),
		"width":      utils.ParseFloat(c.FormValue("width")),
		"height":     utils.ParseFloat(c.FormValue("height")),
		"capturedAt": capturedAt,
	}

	photo, err := h.svc.UploadAndSave(c.UserContext(), file, userId, folderId, meta)
	if err != nil {
		logger.Error("Upload failed: ", err)
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "Upload failed")
	}

	return c.JSON(photo)
}

func (h *PhotoHandler) GetPhoto(c *fiber.Ctx) error {
	id, _ := utils.GetParamAsInt(c, "id")
	folderID, _ := utils.GetParamAsNillableInt(c, "folder_id")

	userId, _ := utils.GetUserIDInt(c)
	item, err := h.svc.GetByID(c.UserContext(), id, userId, folderID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusNotFound, err, "Photo not found")
	}
	return c.JSON(item)
}

func (h *PhotoHandler) GetPhotoByFileName(c *fiber.Ctx) error {
	filename := utils.GetParamAsString(c, "filename")
	folderID, _ := utils.GetParamAsNillableInt(c, "folder_id")

	userId, _ := utils.GetUserIDInt(c)
	item, err := h.svc.GetByFileName(c.UserContext(), filename, userId, folderID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusNotFound, err, "Photo not found")
	}
	return c.JSON(item)
}

func (h *PhotoHandler) GetPhotos(c *fiber.Ctx) error {
	userId := utils.GetQueryAsInt(c, "user_id")
	page := utils.GetQueryAsInt(c, "page", 1)
	limit := utils.GetQueryAsInt(c, "limit", 20)
	folderID, _ := utils.GetQueryAsNillableInt(c, "folder_id")

	list, hasMore, err := h.svc.GetPaginated(c.UserContext(), userId, folderID, page, limit)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "Failed to load photos")
	}
	return c.JSON(fiber.Map{
		"items":   list,
		"hasMore": hasMore,
	})
}

func (h *PhotoHandler) DeletePhoto(c *fiber.Ctx) error {
	id, _ := utils.GetParamAsInt(c, "id")
	folderID, _ := utils.GetParamAsNillableInt(c, "folder_id")
	userId, _ := utils.GetUserIDInt(c)
	err := h.svc.Delete(c.UserContext(), id, userId, folderID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "Delete failed")
	}
	return c.SendStatus(fiber.StatusOK)
}

func (h *PhotoHandler) GetPhotoFile(c *fiber.Ctx) error {
	/*
		GET /api/photo/file/abc.jpg?size=thumbnail
		GET /api/photo/file/abc.jpg?size=medium
		GET /api/photo/file/abc.jpg?size=original
	*/
	filename := c.Params("filename")
	size := c.Query("size", "original") // default: original

	validSizes := map[string]bool{
		"original":  true,
		"medium":    true,
		"thumbnail": true,
	}

	if !validSizes[size] {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "Invalid size parameter")
	}

	basePath := h.deps.Config.Storage.PhotoPath
	basePath = utils.ExpandHomeDir(basePath)

	filePath := filepath.Join(basePath, size, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return client_error.ResponseError(c, fiber.StatusNotFound, err, "File not found")
	}

	return c.SendFile(filePath)
}

type DeleteManyRequest struct {
	IDs      []int `json:"ids"`
	FolderID *int  `json:"folder_id"`
}

func (h *PhotoHandler) DeleteMany(c *fiber.Ctx) error {
	var req DeleteManyRequest
	if err := c.BodyParser(&req); err != nil || len(req.IDs) == 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid request")
	}

	userId, _ := utils.GetUserIDInt(c)

	err := h.svc.DeleteMany(c.UserContext(), req.IDs, userId, req.FolderID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "Delete failed")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *PhotoHandler) BatchGet(c *fiber.Ctx) error {
	var req struct {
		IDs *[]int `json:"ids"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.JSON([]*generated.Photo{})
	}

	if req.IDs == nil || len(*req.IDs) == 0 {
		return c.JSON([]*generated.Photo{})
	}

	userID, _ := utils.GetUserIDInt(c)

	photos, err := h.svc.BatchGetByIDs(c.UserContext(), userID, *req.IDs, nil)

	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "Failed to fetch photos")
	}

	return c.JSON(photos)
}

type UpdateFolderRequest struct {
	IDs         []int `json:"ids"`
	FolderID    *int  `json:"folder_id"`
	OldFolderID *int  `json:"old_folder_id"`
}

func (h *PhotoHandler) UpdateFolder(c *fiber.Ctx) error {
	var req UpdateFolderRequest
	if err := c.BodyParser(&req); err != nil || len(req.IDs) == 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid request")
	}

	userID, _ := utils.GetUserIDInt(c)

	err := h.svc.UpdateFolder(c.UserContext(), userID, req.IDs, req.FolderID, req.OldFolderID)
	if err != nil {
		logger.Error("Failed to update folder for photos: ", err)
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "Update failed")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
