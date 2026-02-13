package history

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/google/uuid"
)

const maxItems = 50

// Step 与 explanation 展示一致
type Step struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	ImageURL string `json:"image_url,omitempty"`
}

// Item 单条历史：上传或文字输入，可选带解析结果
type Item struct {
	ID     string  `json:"id"`
	Type   string  `json:"type"` // "upload" | "text"
	Path   string  `json:"path,omitempty"`
	Text   string  `json:"text,omitempty"`
	At     int64   `json:"at"`
	Result *Result `json:"result,omitempty"`
	TaskID string  `json:"task_id,omitempty"`
}

// Result 解析结果
type Result struct {
	Steps []Step `json:"steps"`
}

// Store 历史存储，内存 + 文件持久化
type Store struct {
	mu       sync.RWMutex
	items    []*Item
	filePath string
}

// NewStore 创建存储，filePath 为空则仅内存
func NewStore(filePath string) (*Store, error) {
	s := &Store{items: make([]*Item, 0), filePath: filePath}
	if filePath != "" {
		if err := s.load(); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		s.mu.Lock()
		s.items = make([]*Item, 0)
		s.mu.Unlock()
		return nil
	}
	var items []*Item
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	if items == nil {
		items = make([]*Item, 0)
	}
	s.mu.Lock()
	s.items = items
	s.mu.Unlock()
	return nil
}

func (s *Store) save() error {
	if s.filePath == "" {
		return nil
	}
	s.mu.RLock()
	snapshot := make([]Item, len(s.items))
	for i, it := range s.items {
		if it != nil {
			snapshot[i] = *it
		}
	}
	s.mu.RUnlock()
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("[history] save mkdir %s: %v", dir, err)
		return err
	}
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		log.Printf("[history] save write %s: %v", s.filePath, err)
		return err
	}
	return nil
}

// List 返回所有历史，新在前，最多 maxItems
func (s *Store) List() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Item, 0, len(s.items))
	for _, it := range s.items {
		out = append(out, *it)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At > out[j].At })
	if len(out) > maxItems {
		out = out[:maxItems]
	}
	return out
}

// Add 新增一条，返回 id
func (s *Store) Add(it Item) string {
	if it.ID == "" {
		it.ID = uuid.New().String()
	}
	s.mu.Lock()
	s.items = append(s.items, &it)
	if len(s.items) > maxItems {
		sort.Slice(s.items, func(i, j int) bool { return s.items[i].At > s.items[j].At })
		s.items = s.items[:maxItems]
	}
	s.mu.Unlock()
	if err := s.save(); err != nil {
		log.Printf("[history] save after Add: %v", err)
	}
	return it.ID
}

// UpdateResult 按 id 更新解析结果
func (s *Store) UpdateResult(id string, result *Result, taskID string) bool {
	s.mu.Lock()
	for _, it := range s.items {
		if it.ID == id {
			it.Result = result
			it.TaskID = taskID
			s.mu.Unlock()
			if err := s.save(); err != nil {
				log.Printf("[history] save after UpdateResult: %v", err)
			}
			return true
		}
	}
	s.mu.Unlock()
	return false
}

// Delete 按 id 删除一条历史
func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, it := range s.items {
		if it != nil && it.ID == id {
			s.items = append(s.items[:i], s.items[i+1:]...)
			if err := s.save(); err != nil {
				log.Printf("[history] save after Delete: %v", err)
			}
			return true
		}
	}
	return false
}

// FindLatestUploadByPath 返回同 path 最新一条（用于当前上传后解析时关联）
func (s *Store) FindLatestUploadByPath(path string) *Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var found *Item
	for _, it := range s.items {
		if it.Type == "upload" && it.Path == path {
			if found == nil || it.At > found.At {
				found = it
			}
		}
	}
	if found == nil {
		return nil
	}
	cp := *found
	return &cp
}
