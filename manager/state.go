package manager

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStateManagement implements StateManagement using Redis
type RedisStateManagement struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisStateManagement creates a new Redis state management instance
func NewRedisStateManagement(addr string) (*RedisStateManagement, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisStateManagement{
		client: client,
		ctx:    ctx,
	}, nil
}

func (s *RedisStateManagement) GetID(name string) string {
	id, _ := s.client.HGet(s.ctx, "tracked", name).Result()
	return id
}

func (s *RedisStateManagement) StoreID(name, id string) error {
	return s.client.HSet(s.ctx, "tracked", name, id).Err()
}

func (s *RedisStateManagement) DestroyID(name string) error {
	id := s.GetID(name)
	if err := s.client.HDel(s.ctx, "tracked", name).Err(); err != nil {
		return err
	}
	return s.client.HDel(s.ctx, "state", id).Err()
}

func (s *RedisStateManagement) SetState(name string, state WindowState) error {
	id := s.GetID(name)
	fmt.Println("setting state: ", id, strconv.Itoa(int(state)))
	return s.client.HSet(s.ctx, "state", id, strconv.Itoa(int(state))).Err()
}

func (s *RedisStateManagement) LatestShown(name string) (string, error) {
	if name != "" {
		err := s.client.ZAdd(s.ctx, "latest", redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: name,
		}).Err()
		return "", err
	}
	result, err := s.client.ZRevRange(s.ctx, "latest", 0, 0).Result()
	if err != nil {
		return "", err
	}
	if len(result) == 0 {
		return "", nil
	}
	return result[0], nil
}

func (s *RedisStateManagement) LatestCount() int {
	count, _ := s.client.ZCount(s.ctx, "latest", "-inf", "+inf").Result()
	return int(count)
}

func (s *RedisStateManagement) IsLatestEmpty() bool {
	return s.LatestCount() == 0
}

func (s *RedisStateManagement) RemoveFromLatest(name string) error {
	return s.client.ZRem(s.ctx, "latest", name).Err()
}

func (s *RedisStateManagement) GetState(id string) WindowState {
	state, _ := s.client.HGet(s.ctx, "state", id).Result()
	stateInt, _ := strconv.Atoi(state)
	return WindowState(stateInt)
}

func (s *RedisStateManagement) IsTracked(name string) bool {
	id := s.GetID(name)
	return id != ""
}

func (s *RedisStateManagement) SaveCurrent(name string, windowType WindowType, focusedID string) error {
	var current string
	if windowType == TypeFocused {
		current = focusedID
	} else if windowType == TypeApplication {
		current = focusedID // This will be replaced by the actual window ID after starting the application
	}
	return s.StoreID(name, current)
}

func (s *RedisStateManagement) StorePrevID(id string) error {
	return s.client.HSet(s.ctx, "tracked", "prev", id).Err()
}

func (s *RedisStateManagement) LoadPrevID() string {
	id, _ := s.client.HGet(s.ctx, "tracked", "prev").Result()
	return id
}

func (s *RedisStateManagement) AllHidden() []struct {
	Name string
	ID   string
} {
	var hidden []struct {
		Name string
		ID   string
	}

	all := s.AllTracked()
	for name, id := range all {
		if s.GetState(id) == NotVisible {
			hidden = append(hidden, struct {
				Name string
				ID   string
			}{name, id})
		}
	}
	return hidden
}

func (s *RedisStateManagement) ResetAll() error {
	if err := s.client.Del(s.ctx, "tracked").Err(); err != nil {
		return err
	}
	return s.client.Del(s.ctx, "state").Err()
}

func (s *RedisStateManagement) AllTracked() map[string]string {
	all, _ := s.client.HGetAll(s.ctx, "tracked").Result()
	return all
}
