package providers

import (
	"github.com/cradio/gormx"
	"github.com/fruitspace/HyprrSpace/models/db"
	"github.com/google/uuid"
)

//region NotificationProvider

type NotificationProvider struct {
	db *gorm.DB
}

func NewNotificationProvider(db *gorm.DB) *NotificationProvider {
	return &NotificationProvider{db: db}
}

func (np *NotificationProvider) GetNotificationsForUID(uid int) []db.Notification {
	var notifications []db.Notification
	np.db.Where(
		np.db.Where(db.Notification{TargetUID: uid}).Or("target_uid=0"),
	).Where(gorm.Column(db.Notification{}, "ExpireDate") + ">CURRENT_TIMESTAMP").
		Order("sendDate DESC").Find(&notifications)
	return notifications
}

func (np *NotificationProvider) GetNotificationByUUID(uuid string) (n db.Notification) {
	np.db.First(&n, uuid)
	return n
}

func (np *NotificationProvider) MarkAsRead(notification db.Notification) {
	np.db.Model(notification).Updates(db.Notification{UserRead: true})
}

func (np *NotificationProvider) CleanExpired() {
	np.db.Where(
		gorm.Column(db.Notification{}, "ExpireDate") + ">CURRENT_TIMESTAMP",
	).Delete(db.Notification{})
}

func (np *NotificationProvider) Delete(uuid string, uid int) {
	np.db.Where(db.Notification{UUID: uuid, TargetUID: uid}).Delete(db.Notification{})
}

func (np *NotificationProvider) Create(notification db.Notification) {
	notification.UUID = uuid.NewString()
	np.db.Create(&notification)
}

//endregion
