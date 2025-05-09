/*
 * WARNING! All changes made in this file will be lost!
 * Created from 'scheme.tl' by 'mtprotoc'
 *
 * Copyright 2025 Teamgram Authors.
 *  All rights reserved.
 *
 * Author: teamgramio (teamgram.io@gmail.com)
 */

package service

import (
	"context"

	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/bff/sponsoredmessages/internal/core"
)

// AccountToggleSponsoredMessages
// account.toggleSponsoredMessages#b9d9a38d enabled:Bool = Bool;
func (s *Service) AccountToggleSponsoredMessages(ctx context.Context, request *mtproto.TLAccountToggleSponsoredMessages) (*mtproto.Bool, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("account.toggleSponsoredMessages - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.AccountToggleSponsoredMessages(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("account.toggleSponsoredMessages - reply: {%s}", r)
	return r, err
}

// ContactsGetSponsoredPeers
// contacts.getSponsoredPeers#b6c8c393 q:string = contacts.SponsoredPeers;
func (s *Service) ContactsGetSponsoredPeers(ctx context.Context, request *mtproto.TLContactsGetSponsoredPeers) (*mtproto.Contacts_SponsoredPeers, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("contacts.getSponsoredPeers - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.ContactsGetSponsoredPeers(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("contacts.getSponsoredPeers - reply: {%s}", r)
	return r, err
}

// MessagesViewSponsoredMessage
// messages.viewSponsoredMessage#269e3643 random_id:bytes = Bool;
func (s *Service) MessagesViewSponsoredMessage(ctx context.Context, request *mtproto.TLMessagesViewSponsoredMessage) (*mtproto.Bool, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("messages.viewSponsoredMessage - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.MessagesViewSponsoredMessage(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("messages.viewSponsoredMessage - reply: {%s}", r)
	return r, err
}

// MessagesClickSponsoredMessage
// messages.clickSponsoredMessage#8235057e flags:# media:flags.0?true fullscreen:flags.1?true random_id:bytes = Bool;
func (s *Service) MessagesClickSponsoredMessage(ctx context.Context, request *mtproto.TLMessagesClickSponsoredMessage) (*mtproto.Bool, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("messages.clickSponsoredMessage - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.MessagesClickSponsoredMessage(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("messages.clickSponsoredMessage - reply: {%s}", r)
	return r, err
}

// MessagesReportSponsoredMessage
// messages.reportSponsoredMessage#12cbf0c4 random_id:bytes option:bytes = channels.SponsoredMessageReportResult;
func (s *Service) MessagesReportSponsoredMessage(ctx context.Context, request *mtproto.TLMessagesReportSponsoredMessage) (*mtproto.Channels_SponsoredMessageReportResult, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("messages.reportSponsoredMessage - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.MessagesReportSponsoredMessage(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("messages.reportSponsoredMessage - reply: {%s}", r)
	return r, err
}

// MessagesGetSponsoredMessages
// messages.getSponsoredMessages#9bd2f439 peer:InputPeer = messages.SponsoredMessages;
func (s *Service) MessagesGetSponsoredMessages(ctx context.Context, request *mtproto.TLMessagesGetSponsoredMessages) (*mtproto.Messages_SponsoredMessages, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("messages.getSponsoredMessages - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.MessagesGetSponsoredMessages(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("messages.getSponsoredMessages - reply: {%s}", r)
	return r, err
}

// ChannelsRestrictSponsoredMessages
// channels.restrictSponsoredMessages#9ae91519 channel:InputChannel restricted:Bool = Updates;
func (s *Service) ChannelsRestrictSponsoredMessages(ctx context.Context, request *mtproto.TLChannelsRestrictSponsoredMessages) (*mtproto.Updates, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("channels.restrictSponsoredMessages - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.ChannelsRestrictSponsoredMessages(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("channels.restrictSponsoredMessages - reply: {%s}", r)
	return r, err
}

// ChannelsViewSponsoredMessage
// channels.viewSponsoredMessage#beaedb94 channel:InputChannel random_id:bytes = Bool;
func (s *Service) ChannelsViewSponsoredMessage(ctx context.Context, request *mtproto.TLChannelsViewSponsoredMessage) (*mtproto.Bool, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("channels.viewSponsoredMessage - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.ChannelsViewSponsoredMessage(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("channels.viewSponsoredMessage - reply: {%s}", r)
	return r, err
}

// ChannelsGetSponsoredMessages
// channels.getSponsoredMessages#ec210fbf channel:InputChannel = messages.SponsoredMessages;
func (s *Service) ChannelsGetSponsoredMessages(ctx context.Context, request *mtproto.TLChannelsGetSponsoredMessages) (*mtproto.Messages_SponsoredMessages, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("channels.getSponsoredMessages - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.ChannelsGetSponsoredMessages(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("channels.getSponsoredMessages - reply: {%s}", r)
	return r, err
}

// ChannelsClickSponsoredMessage
// channels.clickSponsoredMessage#1445d75 flags:# media:flags.0?true fullscreen:flags.1?true channel:InputChannel random_id:bytes = Bool;
func (s *Service) ChannelsClickSponsoredMessage(ctx context.Context, request *mtproto.TLChannelsClickSponsoredMessage) (*mtproto.Bool, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("channels.clickSponsoredMessage - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.ChannelsClickSponsoredMessage(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("channels.clickSponsoredMessage - reply: {%s}", r)
	return r, err
}

// ChannelsReportSponsoredMessage
// channels.reportSponsoredMessage#af8ff6b9 channel:InputChannel random_id:bytes option:bytes = channels.SponsoredMessageReportResult;
func (s *Service) ChannelsReportSponsoredMessage(ctx context.Context, request *mtproto.TLChannelsReportSponsoredMessage) (*mtproto.Channels_SponsoredMessageReportResult, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("channels.reportSponsoredMessage - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.ChannelsReportSponsoredMessage(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("channels.reportSponsoredMessage - reply: {%s}", r)
	return r, err
}
