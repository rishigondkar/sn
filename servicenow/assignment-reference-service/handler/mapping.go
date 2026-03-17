package handler

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/servicenow/assignment-reference-service/proto/assignment_reference_service"
	"github.com/servicenow/assignment-reference-service/repository"
)

func userToProto(u *repository.User) *pb.User {
	if u == nil {
		return nil
	}
	return &pb.User{
		Id:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		IsActive:    u.IsActive,
		CreatedAt:   timestamppb.New(u.CreatedAt),
		UpdatedAt:   timestamppb.New(u.UpdatedAt),
	}
}

func groupToProto(g *repository.Group) *pb.Group {
	if g == nil {
		return nil
	}
	return &pb.Group{
		Id:        g.ID,
		GroupName: g.GroupName,
		Description: g.Description,
		IsActive:  g.IsActive,
		CreatedAt: timestamppb.New(g.CreatedAt),
		UpdatedAt: timestamppb.New(g.UpdatedAt),
	}
}

func groupMemberToProto(m *repository.GroupMember) *pb.GroupMember {
	if m == nil {
		return nil
	}
	return &pb.GroupMember{
		Id:         m.ID,
		GroupId:    m.GroupID,
		UserId:     m.UserID,
		MemberRole: m.MemberRole,
		CreatedAt:  timestamppb.New(m.CreatedAt),
	}
}
