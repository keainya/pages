package object

import "time"

// Page 静态页面模型
type Page struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	URLPath     string    `json:"url_path" gorm:"uniqueIndex;size:255;not null"`
	HTMLContent string    `json:"html_content" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
