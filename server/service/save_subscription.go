package service

import (
	"net/http"

	"github.com/Brightscout/mattermost-plugin-confluence/server/config"
	"github.com/mattermost/mattermost-server/model"

	"github.com/Brightscout/mattermost-plugin-confluence/server/serializer"
	"github.com/Brightscout/mattermost-plugin-confluence/server/store"
	"github.com/pkg/errors"
)

const (
	generalSaveError        = "Failed to save subscription"
	aliasAlreadyExist       = "Subscription already exist in the channel with this alias"
	urlSpaceKeyAlreadyExist = "Subscription already exist in the channel with url and space key combination"
	subscriptionSaveSuccess = "Subscription saved successfully"
)

func SaveNewSubscription(subscription serializer.Subscription, userID string) (int, error) {
	channelSubscriptions, cKey, gErr := GetChannelSubscriptions(subscription.ChannelID)
	if gErr != nil {
		return http.StatusInternalServerError, errors.New(generalSaveError)
	}
	if _, ok := channelSubscriptions[subscription.Alias]; ok {
		return http.StatusBadRequest, errors.New(aliasAlreadyExist)
	}
	keySubscriptions, key, kErr := GetURLSpaceKeyCombinationSubscriptions(subscription.BaseURL, subscription.SpaceKey)
	if kErr != nil {
		return http.StatusInternalServerError, kErr
	}
	if _, ok := keySubscriptions[subscription.ChannelID]; ok {
		return http.StatusBadRequest, errors.New(urlSpaceKeyAlreadyExist)
	}

	keySubscriptions[subscription.ChannelID] = subscription.Events
	channelSubscriptions[subscription.Alias] = subscription
	if err := store.Set(key, keySubscriptions); err != nil {
		return http.StatusInternalServerError, errors.New(generalSaveError)
	}
	if err := store.Set(cKey, channelSubscriptions); err != nil {
		return http.StatusInternalServerError, errors.New(generalSaveError)
	}

	post := &model.Post{
		UserId:    config.BotUserID,
		ChannelId: subscription.ChannelID,
		Message:   subscriptionSaveSuccess,
	}
	_ = config.Mattermost.SendEphemeralPost(userID, post)

	return http.StatusOK, nil
}
