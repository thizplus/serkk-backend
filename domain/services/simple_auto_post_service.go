package services

import (
	"context"
)

// SimpleAutoPostService - บริการโพสต์อัตโนมัติแบบง่าย
type SimpleAutoPostService interface {
	// ProcessNextTopic - ประมวลผล 1 topic ถัดไปจาก queue
	ProcessNextTopic(ctx context.Context) error
}
