package trading

import (
	"context"

	"hotgo/api/admin/trading"
	"hotgo/internal/logic/toogo"
	"hotgo/internal/model/input/toogoin"
)

// StrategyGroup 策略模板控制器
var StrategyGroup = cStrategyGroup{}

type cStrategyGroup struct{}

// List 策略模板列表
func (c *cStrategyGroup) List(ctx context.Context, req *trading.StrategyGroupListReq) (res *trading.StrategyGroupListRes, err error) {
	data, err := toogo.NewStrategyGroupService().List(ctx, &toogoin.StrategyGroupListInp{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Exchange:   req.Exchange,
		Symbol:     req.Symbol,
		IsOfficial: req.IsOfficial,
		NonPersonal: req.NonPersonal,
		IsActive:   req.IsActive,
	})
	if err != nil {
		return nil, err
	}
	res = &trading.StrategyGroupListRes{
		List:  data.List,
		Total: data.Total,
		Page:  data.Page,
	}
	return
}

// Create 创建策略模板
func (c *cStrategyGroup) Create(ctx context.Context, req *trading.StrategyGroupCreateReq) (res *trading.StrategyGroupCreateRes, err error) {
	err = toogo.NewStrategyGroupService().Create(ctx, &toogoin.StrategyGroupCreateInp{
		GroupName:   req.GroupName,
		GroupKey:    req.GroupKey,
		Exchange:    req.Exchange,
		Symbol:      req.Symbol,
		OrderType:   req.OrderType,
		MarginMode:  req.MarginMode,
		Description: req.Description,
		Sort:        req.Sort,
		IsOfficial: req.IsOfficial,
		UserId:     req.UserId,
		IsActive:   req.IsActive,
		IsVisible:  req.IsVisible,
	})
	if err != nil {
		return nil, err
	}
	res = &trading.StrategyGroupCreateRes{}
	return
}

// Update 更新策略模板
func (c *cStrategyGroup) Update(ctx context.Context, req *trading.StrategyGroupUpdateReq) (res *trading.StrategyGroupUpdateRes, err error) {
	err = toogo.NewStrategyGroupService().Update(ctx, &toogoin.StrategyGroupUpdateInp{
		Id:          req.Id,
		GroupName:   req.GroupName,
		Exchange:    req.Exchange,
		Symbol:      req.Symbol,
		OrderType:   req.OrderType,
		MarginMode:  req.MarginMode,
		Description: req.Description,
		Sort:        req.Sort,
		IsVisible:   req.IsVisible,
		IsOfficial:  req.IsOfficial,
		IsActive:    req.IsActive,
		Confirmed:   req.Confirmed,
	})
	if err != nil {
		return nil, err
	}
	res = &trading.StrategyGroupUpdateRes{}
	return
}

// Delete 删除策略模板
func (c *cStrategyGroup) Delete(ctx context.Context, req *trading.StrategyGroupDeleteReq) (res *trading.StrategyGroupDeleteRes, err error) {
	err = toogo.NewStrategyGroupService().Delete(ctx, &toogoin.StrategyGroupDeleteInp{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	res = &trading.StrategyGroupDeleteRes{}
	return
}

// InitStrategies 初始化策略（Init方法名和GoFrame内部冲突，改名）
func (c *cStrategyGroup) InitStrategies(ctx context.Context, req *trading.StrategyGroupInitReq) (res *trading.StrategyGroupInitRes, err error) {
	err = toogo.NewStrategyGroupService().Init(ctx, &toogoin.StrategyGroupInitInp{
		GroupId:    req.GroupId,
		UseDefault: req.UseDefault,
	})
	if err != nil {
		return nil, err
	}
	res = &trading.StrategyGroupInitRes{}
	return
}

// CopyFromOfficial 从官方策略复制到我的策略
func (c *cStrategyGroup) CopyFromOfficial(ctx context.Context, req *trading.StrategyGroupCopyReq) (res *trading.StrategyGroupCopyRes, err error) {
	groupId, err := toogo.NewStrategyGroupService().CopyFromOfficial(ctx, req.OfficialGroupId)
	if err != nil {
		return nil, err
	}
	res = &trading.StrategyGroupCopyRes{
		Id: groupId,
	}
	return
}

// Clone 复制策略组（含策略模板）
func (c *cStrategyGroup) Clone(ctx context.Context, req *trading.StrategyGroupCloneReq) (res *trading.StrategyGroupCloneRes, err error) {
	newId, err := toogo.NewStrategyGroupService().Clone(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	res = &trading.StrategyGroupCloneRes{Id: newId}
	return
}

// SetDefault 设置默认策略模板
func (c *cStrategyGroup) SetDefault(ctx context.Context, req *trading.StrategyGroupSetDefaultReq) (res *trading.StrategyGroupSetDefaultRes, err error) {
	err = toogo.NewStrategyGroupService().SetDefault(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	res = &trading.StrategyGroupSetDefaultRes{}
	return
}
