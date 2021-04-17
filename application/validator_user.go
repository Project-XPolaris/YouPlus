package application

import "github.com/projectxpolaris/youplus/service"

type DuplicateGroupValidator struct {
	Name string
}

func (v *DuplicateGroupValidator) Check() (string, bool) {
	group := service.DefaultUserManager.GetUserGroupByName(v.Name)
	if group != nil {
		return "group already exist!", false
	}
	return "", true
}

type GroupRemoveValidator struct {
	Name string
}

func (v *GroupRemoveValidator) Check() (string, bool) {
	group := service.DefaultUserManager.GetGroupByName(v.Name)
	if group == nil {
		return "group not accessible", false
	}
	if group.Name == service.SuperuserGroup {
		return "admin user group cannot remove!", false
	}
	return "", true
}
