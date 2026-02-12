package explanation

import (
	"sync"

	"github.com/google/uuid"
)

// Store 解析结果存储，供前端与配图步骤使用（当前为内存实现）
type Store struct {
	mu     sync.RWMutex
	byID   map[string]*Result
}

// NewStore 创建内存存储
func NewStore() *Store {
	return &Store{byID: make(map[string]*Result)}
}

// Put 保存解析结果，返回任务 ID
func (s *Store) Put(r *Result) string {
	id := uuid.New().String()
	s.mu.Lock()
	s.byID[id] = r
	s.mu.Unlock()
	return id
}

// Get 按 ID 获取解析结果
func (s *Store) Get(id string) (*Result, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.byID[id]
	return r, ok
}

// Update 更新已存储的解析结果（如写入配图 URL）
func (s *Store) Update(id string, r *Result) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byID[id] = r
}
