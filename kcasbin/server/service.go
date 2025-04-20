package kcasbin

import (
	"context"
	"fmt"

	"github.com/hkensame/goken/pkg/errors"

	"github.com/hkensame/goken/kcasbin/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

var (
	ErrAddPolicyFailed    = errors.New("添加策略失败")
	ErrRemovePolicyFailed = errors.New("删除策略失败")
	ErrUpdatePolicyFailed = errors.New("更新策略失败")
	ErrFindPolicyFailed   = errors.New("查找策略失败")

	ErrAuthorizedFailed = errors.New("鉴权失败")
)

func readPolicies(in *proto.MatchPolicies) [][]string {
	pc := [][]string{}
	for _, v := range in.Mp {
		pc = append(pc, []string{v.Sub, v.Dom, v.Obj, v.Eft})
	}
	return pc
}

func readGroupingPolicies(in *proto.GroupingPolicies) [][]string {
	pc := [][]string{}
	for _, v := range in.Gp {
		pc = append(pc, []string{v.Rsub, v.Psub, v.Dom})
	}
	return pc
}

func (s *Kasbin) AddPolicies(ctx context.Context, in *proto.MatchPolicies) (*emptypb.Empty, error) {
	if _, err := s.Casb.AddNamedPolicies(in.Pname, readPolicies(in)); err != nil {
		s.Logger.Sugar().Errorf("[kcasbin] 策略添加失败, err = %v", err)
		return nil, ErrAddPolicyFailed
	}
	return &emptypb.Empty{}, nil
}

func (s *Kasbin) RemovePolicies(ctx context.Context, in *proto.MatchPolicies) (*emptypb.Empty, error) {
	if _, err := s.Casb.RemoveNamedPolicy(in.Pname, readPolicies(in)); err != nil {
		s.Logger.Sugar().Errorf("[kcasbin] 策略删除失败, err = %v", err)
		return nil, ErrRemovePolicyFailed
	}
	return &emptypb.Empty{}, nil
}

func (s *Kasbin) UpdatePolicies(ctx context.Context, in *proto.UpdatePoliciesReq) (*emptypb.Empty, error) {
	old := readPolicies(in.OldPolicies)
	new := readPolicies(in.OldPolicies)
	if _, err := s.Casb.UpdateNamedPolicies(in.Pname, old, new); err != nil {
		s.Logger.Sugar().Errorf("[kcasbin] 策略更新失败, err = %v", err)
		return nil, ErrRemovePolicyFailed
	}
	return &emptypb.Empty{}, nil
}

// 后续这个函数一定要改,要参数来过滤查找得到的policy
func (s *Kasbin) FindPoliciesItems(ctx context.Context, in *proto.FindPoliciesFilterReq) (*proto.MatchPolicies, error) {
	policies, err := s.Casb.GetNamedPolicy(in.Pname)
	if err != nil {
		s.Logger.Sugar().Errorf("[kcasbin] 策略查找失败, err = %v", err)
		return nil, ErrFindPolicyFailed
	}
	res := &proto.MatchPolicies{}
	for _, p := range policies {
		if len(p) < 4 {
			continue
		}
		res.Mp = append(res.Mp, &proto.MatchPolicy{
			Pname: in.Pname,
			Sub:   p[0],
			Dom:   p[1],
			Obj:   p[2],
			Eft:   p[3],
		})
	}
	return res, nil
}

func (s *Kasbin) AddGroupingPolicies(ctx context.Context, in *proto.GroupingPolicies) (*emptypb.Empty, error) {
	if _, err := s.Casb.AddNamedPolicies(in.Gname, readGroupingPolicies(in)); err != nil {
		s.Logger.Sugar().Errorf("[kcasbin] 策略添加失败, err = %v", err)
		return nil, ErrAddPolicyFailed
	}
	return &emptypb.Empty{}, nil
}

func (s *Kasbin) RemoveGroupingPolicies(ctx context.Context, in *proto.GroupingPolicies) (*emptypb.Empty, error) {
	if _, err := s.Casb.RemoveNamedPolicy(in.Gname, readGroupingPolicies(in)); err != nil {
		s.Logger.Sugar().Errorf("[kcasbin] 策略删除失败, err = %v", err)
		return nil, ErrRemovePolicyFailed
	}
	return &emptypb.Empty{}, nil
}

// 同样要实现一致性
func (s *Kasbin) UpdateGroupingPolicies(ctx context.Context, in *proto.UpdateGroupingPoliciesReq) (*emptypb.Empty, error) {
	old := readGroupingPolicies(in.OldPolicies)
	new := readGroupingPolicies(in.NewPolicies)
	if _, err := s.Casb.UpdateNamedPolicies(in.Gname, old, new); err != nil {
		s.Logger.Sugar().Errorf("[kcasbin] 策略更新失败, err = %v", err)
		return nil, ErrRemovePolicyFailed
	}
	return &emptypb.Empty{}, nil
}

func (s *Kasbin) FindGroupingPoliciesItems(ctx context.Context, in *proto.FindGroupingPoliciesFilterReq) (*proto.GroupingPolicies, error) {
	policies, err := s.Casb.GetNamedPolicy(in.Gname)
	if err != nil {
		s.Logger.Sugar().Errorf("[kcasbin] 策略查找失败, err = %v", err)
		return nil, ErrFindPolicyFailed
	}
	res := &proto.GroupingPolicies{}
	for _, p := range policies {
		if len(p) < 3 {
			continue
		}
		res.Gp = append(res.Gp, &proto.GroupingPolicy{
			Gname: in.Gname,
			Rsub:  p[0],
			Psub:  p[1],
			Dom:   p[2],
		})
	}
	return res, nil
}

func (s *Kasbin) Authorize(ctx context.Context, in *proto.AuthorizeReq) (*proto.AuthorizeRes, error) {
	ok, err := s.Casb.Enforce(in.Sub, in.Dom, in.Obj)
	res := &proto.AuthorizeRes{
		Ok: ok,
	}
	if err != nil || !ok {
		if err != nil {
			s.Logger.Sugar().Errorf("[kcasbin] 授权失败: %v", err)
		}
		res.Detail = fmt.Sprintf("用户[%s] 对资源[%s:%s]授权失败", in.Sub, in.Dom, in.Obj)
		return res, ErrAuthorizedFailed
	}
	res.Detail = fmt.Sprintf("用户[%s] 对资源[%s:%s] 授权结果: %v", in.Sub, in.Dom, in.Obj, ok)
	return res, nil
}

func (s *Kasbin) GetUserRoles(ctx context.Context, in *proto.GetUserRolesReq) (*proto.GetUserRolesRes, error) {
	roles, err := s.Casb.GetRolesForUser(in.User)
	if err != nil {
		return nil, fmt.Errorf("获取角色失败: %v", err)
	}
	return &proto.GetUserRolesRes{Roles: roles}, nil
}

func (s *Kasbin) GetRoleUsers(ctx context.Context, in *proto.GetRoleUsersReq) (*proto.GetRoleUsersRes, error) {
	users, err := s.Casb.GetUsersForRole(in.Role)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %v", err)
	}
	return &proto.GetRoleUsersRes{Users: users}, nil
}
