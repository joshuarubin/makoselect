package main

import "fmt"

type NotificationID struct {
	Data int64
}

type NotificationActions struct {
	Data map[string]string
}

type Notification struct {
	ID      *NotificationID
	Actions *NotificationActions
}

func (n *Notification) Is(id int64) bool {
	if n == nil || n.ID == nil || n.ID.Data != id || n.Actions == nil || n.Actions.Data == nil {
		return false
	}
	return true
}

type Notifications struct {
	Data [][]*Notification
}

func (n *Notifications) GetByID(id int64) (*Notification, error) {
	if n == nil || n.Data == nil {
		return nil, fmt.Errorf("no notifications")
	}

	for _, list := range n.Data {
		for _, item := range list {
			if item.Is(id) {
				return item, nil
			}
		}
	}

	return nil, fmt.Errorf("notification not found")
}
