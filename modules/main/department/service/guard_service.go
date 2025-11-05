package service

import (
	"context"
	"fmt"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/department/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/middleware"
)

type guardService struct {
	repo repository.DepartmentRepository
}

func NewGuardService(repo repository.DepartmentRepository) middleware.DepartmentChecker {
	return &guardService{repo: repo}
}

func keyGuardMember(u, d int) string { return fmt.Sprintf("department:member:u%d:d%d", u, d) }

func (s *guardService) IsMember(ctx context.Context, userID, deptID int) (bool, error) {
	type R struct{ OK bool }
	r, err := cache.Get(keyGuardMember(userID, deptID), 2*time.Minute, func() (*R, error) {
		ok, err := s.repo.ExistsMembership(ctx, userID, deptID)
		if err != nil {
			return nil, err
		}
		return &R{OK: ok}, nil
	})
	if err != nil {
		return false, err
	}
	return r.OK, nil
}
