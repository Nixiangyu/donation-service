/*
Copyright Lingzhu Ltd. 2020 All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package image

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/csiabb/donation-service/common/rest"
	"github.com/csiabb/donation-service/common/utils"
	"github.com/csiabb/donation-service/structs"

	"github.com/gin-gonic/gin"
)

// Upload defines the upload of image
func (h *RestHandler) Upload(c *gin.Context) {
	logger.Info("got image upload request")

	c.Request.ParseMultipartForm(32 << 20)
	fileRec, _, err := c.Request.FormFile("image_file")
	if err != nil {
		e := fmt.Errorf("invalid parameters, %s", err.Error())
		logger.Error(e)
		c.JSON(http.StatusBadRequest, rest.ErrorResponse(rest.InvalidParamsErrCode, e.Error()))
		return
	}
	logger.Debugf("request params %v", fileRec)

	imageFileTag := utils.GenerateUUID()

	if err = h.srvcContext.ALiYunBackend.UploadObject(imageFileTag+".png", fileRec); err != nil {
		e := fmt.Errorf("image upload error : %s", err.Error())
		logger.Error(e)
		c.JSON(http.StatusBadRequest, rest.ErrorResponse(rest.InternalServerFailure, e.Error()))
		return
	}

	c.JSON(http.StatusOK, rest.SuccessResponse(&structs.ImageUploadResp{
		ID: imageFileTag,
	}))

	logger.Info("response image upload success.")
	return
}

// Draw define draw of image
func (h *RestHandler) Draw(c *gin.Context) {
	logger.Infof("Got query share request")

	req := &structs.DrawRequest{}
	if err := c.Bind(req); err != nil {
		e := fmt.Errorf("invalid parameters: %s", err.Error())
		logger.Error(e)
		c.JSON(http.StatusBadRequest, rest.ErrorResponse(rest.ParseRequestParamsError, e.Error()))
		return
	}
	logger.Debugf("request params, %v", req)

	var err error
	var content, imagURL string
	if req.DrawType == rest.Prove {
		content = rest.DrawDonationContent
		if imagURL, err = h.GetImageURL(req); err != nil {
			e := fmt.Errorf("get image url err : %s", err.Error())
			logger.Error(e)
			c.JSON(http.StatusBadRequest, rest.ErrorResponse(rest.InternalServerFailure, e.Error()))
			return
		}
	} else {
		content = rest.DrawHomeContent
		imagURL = h.srvcContext.Config.LocalFileSystem + rest.DrawImageURL
	}

	c.JSON(http.StatusOK, rest.SuccessResponse(&structs.DrawResp{
		Icon:     h.srvcContext.Config.LocalFileSystem + rest.DrawIcon,
		Title:    rest.DrawTitle,
		Content:  content,
		ImageURL: imagURL,
	}))

	logger.Info("response share success.")
	return

}

//  Download defines the download of image
func (h *RestHandler) Download(c *gin.Context) {
	logger.Info("got image download request")
	req := &structs.DownloadRequest{}
	if err := c.Bind(req); err != nil {
		e := fmt.Errorf("invalid parameters: %s", err.Error())
		logger.Error(e)
		c.JSON(http.StatusBadRequest, rest.ErrorResponse(rest.ParseRequestParamsError, e.Error()))
		return
	}
	logger.Debugf("request params, %v", req)

	reader, e := h.srvcContext.ALiYunBackend.DownloadObject(req.ImageUrl)
	if e != nil {
		logger.Error(e)
		c.JSON(http.StatusBadRequest, rest.ErrorResponse(rest.InternalServerFailure, e.Error()))
		return
	} else if reader.Response.StatusCode != 200 {
		e = errors.New("download is failed , code is not equal 200")
		logger.Error(e)
		c.JSON(http.StatusBadRequest, rest.ErrorResponse(rest.InternalServerFailure, e.Error()))
		return
	}

	res, e := ioutil.ReadAll(reader.Response.Body)
	if e != nil {
		logger.Error(e)
		c.JSON(http.StatusBadRequest, rest.ErrorResponse(rest.InternalServerFailure, e.Error()))
		return
	}

	c.Writer.Write(res)
	logger.Info("response image download success.")
	return
}
