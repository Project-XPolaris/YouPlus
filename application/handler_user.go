package application

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/validator"
	"github.com/pkg/errors"
	"github.com/projectxpolaris/youplus/service"
	"net/http"
)

type CreateUserRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var createUserHandler haruka.RequestHandler = func(context *haruka.Context) {
	var body CreateUserRequestBody
	err := context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.DefaultUserManager.NewUser(body.Username, body.Password, false)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var removeUserHandler haruka.RequestHandler = func(context *haruka.Context) {
	username := context.GetQueryString("username")
	err := service.DefaultUserManager.RemoveUser(username)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var getUserList haruka.RequestHandler = func(context *haruka.Context) {
	userList, err := service.GetUserList()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"users": userList,
	})
}

type UserAuthRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var userLoginHandler haruka.RequestHandler = func(context *haruka.Context) {
	var body UserAuthRequestBody
	err := context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	uid, tokenStr, err := service.UserLogin(body.Username, body.Password, true)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
		"token":   tokenStr,
		"uid":     uid,
	})
}
var generateAuthHandler haruka.RequestHandler = func(context *haruka.Context) {
	var body UserAuthRequestBody
	err := context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	uid, tokenStr, err := service.UserLogin(body.Username, body.Password, false)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
		"token":   tokenStr,
		"uid":     uid,
	})
}
var checkTokenHandler haruka.RequestHandler = func(context *haruka.Context) {
	rawToken := context.GetQueryString("token")
	user, err := service.ParseUser(rawToken)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"success":  true,
		"username": user.Username,
		"uid":      user.Uid,
	})
}

var userGroupListHandler haruka.RequestHandler = func(context *haruka.Context) {
	groups := make([]UserGroupTemplate, 0)
	sysGroups, err := service.DefaultUserManager.GetGroups()
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	for _, group := range sysGroups {
		template := UserGroupTemplate{}
		template.Assign(group)
		groups = append(groups, template)
	}
	context.JSON(haruka.JSON{
		"groups": groups,
	})
}

type CreateUserGroupRequestBody struct {
	Name string `json:"name"`
}

var addUserGroup haruka.RequestHandler = func(context *haruka.Context) {
	var body CreateUserGroupRequestBody
	err := context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	if err = validator.RunValidators(&DuplicateGroupValidator{Name: body.Name}); err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	group, err := service.DefaultUserManager.CreateGroup(body.Name)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	template := UserGroupTemplate{}
	template.Assign(group)
	context.JSON(template)
}

var removeUserGroup haruka.RequestHandler = func(context *haruka.Context) {
	name := context.GetQueryString("name")
	if err := validator.RunValidators(&GroupRemoveValidator{Name: name}); err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err := service.DefaultUserManager.RemoveUserGroup(name)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

type UserGroupInput struct {
	Name string `hsource:"path" hname:"name"`
}

var userGroupHandler haruka.RequestHandler = func(context *haruka.Context) {
	var input UserGroupInput
	err := context.BindingInput(&input)
	if err != nil {
		AbortErrorWithStatus(err, context, 400)
		return
	}
	group := service.DefaultUserManager.GetGroupByName(input.Name)
	if group == nil {
		AbortErrorWithStatus(errors.New("group not found"), context, http.StatusNotFound)
		return
	}
	template := UserGroupTemplate{}
	template.Assign(group)
	var users []*service.SystemUser
	linq.From(service.DefaultUserManager.Users).Where(func(i interface{}) bool {
		for _, name := range group.Users {
			if name == i.(*service.SystemUser).Username {
				return true
			}
		}
		return false
	}).ToSlice(&users)
	usersTemplates := make([]UserTemplate, 0)
	for _, systemUser := range users {
		userTemplate := UserTemplate{}
		userTemplate.Uid = systemUser.Uid
		userTemplate.Name = systemUser.Username
		usersTemplates = append(usersTemplates, userTemplate)
	}
	template.Users = usersTemplates
	context.JSON(template)
}

type UserGroupUserActionRequestBody struct {
	Users []string `json:"users"`
}

var removeUserFromUserGroupHandler haruka.RequestHandler = func(context *haruka.Context) {
	var input UserGroupInput
	err := context.BindingInput(&input)
	if err != nil {
		AbortErrorWithStatus(err, context, 400)
		return
	}
	var body UserGroupUserActionRequestBody
	err = context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	group := service.DefaultUserManager.GetGroupByName(input.Name)
	if group == nil {
		AbortErrorWithStatus(errors.New("group not found"), context, http.StatusNotFound)
		return
	}
	for _, username := range body.Users {
		user := service.DefaultUserManager.GetUserByName(username)
		if user == nil {
			continue
		}
		err = group.DelUser(user)
		if err != nil {
			AbortErrorWithStatus(err, context, 500)
			return
		}
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
var addUserToUserGroupHandler haruka.RequestHandler = func(context *haruka.Context) {
	var input UserGroupInput
	err := context.BindingInput(&input)
	if err != nil {
		AbortErrorWithStatus(err, context, 400)
		return
	}
	var body UserGroupUserActionRequestBody
	err = context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	group := service.DefaultUserManager.GetGroupByName(input.Name)
	if group == nil {
		AbortErrorWithStatus(errors.New("group not found"), context, http.StatusNotFound)
		return
	}
	for _, username := range body.Users {
		user := service.DefaultUserManager.GetUserByName(username)
		if user == nil {
			continue
		}
		err = group.AddUser(user)
		if err != nil {
			AbortErrorWithStatus(err, context, 500)
			return
		}
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
