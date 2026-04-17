package model

type ProductEventConsumedModel struct {
	ID          int64  `gorm:"primaryKey;autoIncrement"`
	EventSource string `gorm:"type:varchar(128);not null;uniqueIndex:uk_event_msg,priority:1"`
	EventId     string `gorm:"type:varchar(128);not null;uniqueIndex:uk_event_msg,priority:2"`
	ProductId   int64  `gorm:"not null"`
	Count       int64  `gorm:"not null"`
	EventType   string `gorm:"type:varchar(128);not null"`
	CreatedAt   int64  `gorm:"not null"`
}

func (ProductEventConsumedModel) TableName() string {
	return "product_event_consumed"
}
