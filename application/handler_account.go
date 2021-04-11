package application

import (
	"github.com/allentom/haruka"
	"github.com/dgrijalva/jwt-go"
	"github.com/projectxpolaris/youplus/service"
)

type ChangePasswordInput struct {
	Claims *jwt.StandardClaims `hsource:"param" hname:"claims"`
}
type ChangePasswordForm struct {
	Password string `json:"password"`
}

var changeAccountPasswordHandler haruka.RequestHandler = func(context *haruka.Context) {
	var input ChangePasswordInput
	err := context.BindingInput(&input)
	if err != nil {
		AbortErrorWithStatus(err, context, 400)
		return
	}
	var form ChangePasswordForm
	err = context.ParseJson(&form)
	if err != nil {
		AbortErrorWithStatus(err, context, 400)
		return
	}
	err = service.DefaultUserManager.ChangeUserPassword(input.Claims.Id, form.Password)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
