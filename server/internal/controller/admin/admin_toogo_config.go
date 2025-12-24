// Package admin Toogo系统配置控制器
package admin

import (
	"context"

	"hotgo/api/admin"
	"hotgo/internal/logic/toogo"
)

var ToogoConfig = cToogoConfig{}

type cToogoConfig struct{}

// Groups 获取配置分组
func (c *cToogoConfig) Groups(ctx context.Context, req *admin.ToogoConfigGroupsReq) (res *admin.ToogoConfigGroupsRes, err error) {
	groups := toogo.GetConfig().GetGroups(ctx)
	result := make([]admin.ConfigGroup, len(groups))
	for i, g := range groups {
		result[i] = admin.ConfigGroup{Key: g.Key, Label: g.Label}
	}
	res = &admin.ToogoConfigGroupsRes{
		Groups: result,
	}
	return
}

// List 获取配置列表
func (c *cToogoConfig) List(ctx context.Context, req *admin.ToogoConfigListReq) (res *admin.ToogoConfigListRes, err error) {
	list, err := toogo.GetConfig().GetList(ctx, req.Group)
	if err != nil {
		return nil, err
	}

	items := make([]*admin.ToogoConfigItem, 0, len(list))
	for _, item := range list {
		items = append(items, &admin.ToogoConfigItem{
			Id:          item.Id,
			Group:       item.Group,
			Key:         item.Key,
			Value:       item.Value,
			Type:        item.Type,
			Name:        item.Name,
			Description: item.Description,
			Sort:        item.Sort,
		})
	}

	res = &admin.ToogoConfigListRes{List: items}
	return
}

// Get 获取单个配置
func (c *cToogoConfig) Get(ctx context.Context, req *admin.ToogoConfigGetReq) (res *admin.ToogoConfigGetRes, err error) {
	item, err := toogo.GetConfig().Get(ctx, req.Group, req.Key)
	if err != nil {
		return nil, err
	}

	if item == nil {
		res = &admin.ToogoConfigGetRes{}
		return
	}

	res = &admin.ToogoConfigGetRes{
		ToogoConfigItem: &admin.ToogoConfigItem{
			Id:          item.Id,
			Group:       item.Group,
			Key:         item.Key,
			Value:       item.Value,
			Type:        item.Type,
			Name:        item.Name,
			Description: item.Description,
			Sort:        item.Sort,
		},
	}
	return
}

// Update 更新配置
func (c *cToogoConfig) Update(ctx context.Context, req *admin.ToogoConfigUpdateReq) (res *admin.ToogoConfigUpdateRes, err error) {
	items := make([]toogo.ConfigUpdateItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, toogo.ConfigUpdateItem{
			Group: item.Group,
			Key:   item.Key,
			Value: item.Value,
		})
	}

	err = toogo.GetConfig().Update(ctx, items)
	if err != nil {
		return nil, err
	}

	res = &admin.ToogoConfigUpdateRes{}
	return
}

