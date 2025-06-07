package controller

import (
	"encoding/json"
	"fido2/pkg/utils/common"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// AppleWellKnownHandler handles requests to the /.well-known/apple-app-site-association endpoint.
func AppleWellKnownHandler(ctx *gin.Context) {
	fmt.Println("call /.well-known/apple-app-site-association")

	appleAppSiteAssociationData, err := os.ReadFile("apple-app-site-association")
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to get apple-app-site-association",
			},
		)
		return
	}

	var appleAppSiteAssociation map[string]interface{}
	if err = json.Unmarshal(appleAppSiteAssociationData, &appleAppSiteAssociation); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to parse apple-app-site-association",
			},
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		appleAppSiteAssociation,
	)
}

func AndroidWellKnownHandler(ctx *gin.Context) {
	fmt.Println("call /.well-known/assetlinks.json")

	assetlinksData, err := os.ReadFile("assetlinks.json")
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to get assetlinks.json",
			},
		)
		return
	}

	var assetlinks []map[string]interface{}
	if err = json.Unmarshal(assetlinksData, &assetlinks); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to parse assetlinks.json",
			},
		)
		return
	}
	ctx.JSON(
		http.StatusOK,
		assetlinks,
	)
}
