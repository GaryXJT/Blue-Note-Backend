package service

import (
	"sync"
	"time"

	"github.com/mojocn/base64Captcha"
)

// CaptchaService 验证码服务
type CaptchaService struct {
	store         *customStore
	captchaWidth  int
	captchaHeight int
	captchaLength int
}

// customStore 自定义验证码存储
type customStore struct {
	sync.RWMutex
	data map[string]item
}

// item 验证码项
type item struct {
	code     string
	expireAt time.Time
}

// Set 存储验证码
func (s *customStore) Set(id string, value string) error {
	s.Lock()
	defer s.Unlock()
	s.data[id] = item{
		code:     value,
		expireAt: time.Now().Add(5 * time.Minute), // 验证码 5 分钟内有效
	}
	return nil
}

// Get 获取验证码
func (s *customStore) Get(id string, clear bool) string {
	s.RLock()
	defer s.RUnlock()
	if item, ok := s.data[id]; ok {
		if item.expireAt.After(time.Now()) {
			if clear {
				delete(s.data, id)
			}
			return item.code
		}
		// 已过期
		delete(s.data, id)
	}
	return ""
}

// Verify 验证验证码
func (s *customStore) Verify(id, answer string, clear bool) bool {
	v := s.Get(id, clear)
	return v == answer
}

// NewCaptchaService 创建一个新的验证码服务
func NewCaptchaService() *CaptchaService {
	// 创建自定义存储
	store := &customStore{
		data: make(map[string]item),
	}

	// 定期清理过期验证码（每10分钟执行一次）
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			store.Lock()
			now := time.Now()
			for id, item := range store.data {
				if item.expireAt.Before(now) {
					delete(store.data, id)
				}
			}
			store.Unlock()
		}
	}()

	return &CaptchaService{
		store:         store,
		captchaWidth:  240,
		captchaHeight: 80,
		captchaLength: 4,
	}
}

// GenerateCaptcha 生成验证码
func (s *CaptchaService) GenerateCaptcha() (string, string, error) {
	// 创建驱动配置
	driver := base64Captcha.NewDriverDigit(
		s.captchaHeight,
		s.captchaWidth,
		s.captchaLength,
		0.7,
		5,
	)

	// 创建验证码对象
	captcha := base64Captcha.NewCaptcha(driver, s.store)

	// 生成验证码
	id, b64s, _, err := captcha.Generate()
	if err != nil {
		return "", "", err
	}

	return id, b64s, nil
}

// VerifyCaptcha 验证验证码
func (s *CaptchaService) VerifyCaptcha(id, answer string) bool {
	return s.store.Verify(id, answer, true)
}
