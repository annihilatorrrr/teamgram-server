/*
 * Created from 'scheme.tl' by 'mtprotoc'
 *
 * Copyright (c) 2021-present,  Teamgram Studio (https://teamgram.io).
 *  All rights reserved.
 *
 * Author: teamgramio (teamgram.io@gmail.com)
 */

package core

import (
	"context"

	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/messenger/msg/inbox/inbox"
	"github.com/teamgram/teamgram-server/app/messenger/msg/msg/msg"
	"github.com/teamgram/teamgram-server/app/messenger/msg/msg/plugin"
	"github.com/teamgram/teamgram-server/app/messenger/sync/sync"
	chatpb "github.com/teamgram/teamgram-server/app/service/biz/chat/chat"
	userpb "github.com/teamgram/teamgram-server/app/service/biz/user/user"

	"github.com/zeromicro/go-zero/core/mr"
)

// MsgSendMessage
// msg.sendMessage user_id:long auth_key_id:long peer_type:int peer_id:long message:OutboxMessage = Updates;
func (c *MsgCore) MsgSendMessage(in *msg.TLMsgSendMessage) (*mtproto.Updates, error) {
	var (
		rUpdates *mtproto.Updates
		err      error
		outBox   = in.GetMessage()
		peer     = mtproto.MakePeerUtil(in.PeerType, in.PeerId)
	)

	if peer.IsChannel() {
		// c.Logger.Errorf("msg.sendMultiMessage blocked, License key from https://teamgram.net required to unlock enterprise features.")
		return nil, mtproto.ErrEnterpriseIsBlocked
	}

	if outBox.GetScheduleDate().GetValue() != 0 {
		// c.Logger.Errorf("msg.sendMessage blocked, License key from https://teamgram.net required to unlock enterprise features.")
		return nil, mtproto.ErrEnterpriseIsBlocked
	}

	if !peer.IsChatOrUser() {
		err = mtproto.ErrPeerIdInvalid
		c.Logger.Errorf("msg.sendMessage - error: %v", err)
		return nil, err
	}

	if peer.IsUser() {
		rUpdates, err = c.sendUserOutgoingMessageV2(in.UserId, in.AuthKeyId, in.PeerId, outBox)
		if err != nil {
			c.Logger.Errorf("msg.sendMessage - error: %v", err)
			return nil, err
		}
	} else {
		rUpdates, err = c.sendChatOutgoingMessageV2(in.UserId, in.AuthKeyId, in.PeerId, outBox)
		if err != nil {
			c.Logger.Errorf("msg.sendMessage - error: %v", err)
			return nil, err
		}
	}

	return rUpdates, nil
}

func (c *MsgCore) sendUserOutgoingMessage(userId, authKeyId, peerUserId int64, outBox *msg.OutboxMessage) (*mtproto.Updates, error) {
	var (
		err      error
		rUpdates *mtproto.Updates
	)

	rUpdates, err = c.sendUserMessage(
		c.ctx,
		userId,
		authKeyId,
		peerUserId,
		outBox,
		func(did int64, inboxMsg *mtproto.Message) error {
			blocked, _ := c.svcCtx.Dao.UserClient.UserBlockedByUser(c.ctx, &userpb.TLUserBlockedByUser{
				UserId:     peerUserId,
				PeerUserId: userId,
			})
			if !mtproto.FromBool(blocked) {
				_, err = c.svcCtx.Dao.InboxClient.InboxSendUserMessageToInbox(c.ctx, &inbox.TLInboxSendUserMessageToInbox{
					FromId:     userId,
					PeerUserId: peerUserId,
					Message: inbox.MakeTLInboxMessageData(&inbox.InboxMessageData{
						RandomId:        outBox.RandomId,
						DialogMessageId: did,
						// MessageDataId:   mid,
						Message: inboxMsg,
					}).To_InboxMessageData(),
				})
			}
			return nil
		})
	if err != nil {
		c.Logger.Errorf("msg.sendUserOutgoingMessage - error: %v", err)
		return nil, err
	}

	return rUpdates, nil
}

func (c *MsgCore) sendUserMessage(
	ctx context.Context,
	fromUserId int64,
	fromAuthKeyId int64,
	toUserId int64,
	outBox *msg.OutboxMessage,
	cb func(did int64, inboxMsg *mtproto.Message) error,
) (*mtproto.Updates, error) {
	users, err := c.svcCtx.Dao.UserClient.UserGetMutableUsers(ctx, &userpb.TLUserGetMutableUsers{
		Id: []int64{fromUserId, toUserId},
	})
	if err != nil {
		c.Logger.Errorf("msg.sendUserOutgoingMessage - error: %v", err)
		return nil, err
	}

	sender, _ := users.GetImmutableUser(fromUserId)
	if sender == nil || sender.Deleted() {
		err = mtproto.ErrInputUserDeactivated
		c.Logger.Errorf("msg.sendUserOutgoingMessage - error: %v", err)
		return nil, err
	}

	// TODO(@benqi): check
	// if sender.Restricted() {
	//	err = mtproto.ErrUserRestricted
	//	return
	// }

	peerUser, _ := users.GetImmutableUser(toUserId)
	if peerUser == nil || peerUser.Deleted() {
		err = mtproto.ErrInputUserDeactivated
		c.Logger.Errorf("msg.sendUserOutgoingMessage - error: %v", err)
		return nil, err
	}

	sendMe := fromUserId == toUserId
	if !sendMe {
		// TODO(@benqi)
		// 1. check blocked
		// 2. span
	}

	var (
		rUpdates         *mtproto.Updates
		notSYncMe        = false
		updateNewMessage *mtproto.Update
	)

	cached, err := c.svcCtx.Dao.DoIdempotent(
		ctx,
		fromUserId,
		outBox.RandomId,
		&rUpdates,
		func(ctx context.Context, v any) error {
			box, ok, err := c.svcCtx.Dao.SendUserMessage(ctx, fromUserId, toUserId, outBox)
			if err != nil {
				c.Logger.Error(err.Error())
				return err
			}

			if ok && cb != nil {
				err = cb(box.DialogMessageId, box.ToMessage(fromUserId))
				if err != nil {
					c.Logger.Error(err.Error())
					return err
				}
			}

			updateNewMessage = mtproto.MakeTLUpdateNewMessage(&mtproto.Update{
				Pts_INT32:       box.Pts,
				PtsCount:        box.PtsCount,
				RandomId:        box.RandomId,
				Message_MESSAGE: box.Message,
			}).To_Update()

			*v.(**mtproto.Updates) = mtproto.MakeReplyUpdates(
				func(idList []int64) []*mtproto.User {
					// TODO: check
					//users, _ := c.svcCtx.Dao.UserClient.UserGetMutableUsers(ctx,
					//	&userpb.TLUserGetMutableUsers{
					//		Id: idList,
					//	})
					return users.GetUserListByIdList(fromUserId, idList...)
				},
				func(idList []int64) []*mtproto.Chat {
					if len(idList) > 0 {
						chats, _ := c.svcCtx.Dao.ChatClient.ChatGetChatListByIdList(ctx,
							&chatpb.TLChatGetChatListByIdList{
								IdList: idList,
							})
						return chats.GetChatListByIdList(fromUserId, idList...)
					} else {
						return []*mtproto.Chat{}
					}
				},
				func(idList []int64) []*mtproto.Chat {
					// TODO
					return nil
				},
				updateNewMessage)

			notSYncMe = !ok
			if !ok {
				// dup
				return nil
			}

			// c.svcCtx.Dao.MessageDeDuplicate.PutDuplicateMessage(ctx, fromUserId, outBox.RandomId, rUpdates)

			return nil
		})
	if err == nil {
		if !cached && notSYncMe {
			c.svcCtx.Dao.SyncClient.SyncUpdatesNotMe(ctx, &sync.TLSyncUpdatesNotMe{
				UserId:        fromUserId,
				PermAuthKeyId: fromAuthKeyId,
				Updates: mtproto.MakeSyncNotMeUpdates(
					func(idList []int64) []*mtproto.User {
						return rUpdates.Users
					},
					func(idList []int64) []*mtproto.Chat {
						return rUpdates.Chats
					},
					func(idList []int64) []*mtproto.Chat {
						// rUpdates.Chats include chats
						return nil
					},
					updateNewMessage),
			})
		}
	}

	return rUpdates, err
}

func (c *MsgCore) sendChatOutgoingMessage(userId, authKeyId, peerChatId int64, outBox *msg.OutboxMessage) (*mtproto.Updates, error) {
	rUpdates, err := c.sendChatMessage(c.ctx,
		userId,
		authKeyId,
		peerChatId,
		outBox,
		func(did int64, inboxMsg *mtproto.Message) error {
			_, err := c.svcCtx.Dao.InboxClient.InboxSendChatMessageToInbox(
				c.ctx,
				&inbox.TLInboxSendChatMessageToInbox{
					FromId:     userId,
					PeerChatId: peerChatId,
					Message: inbox.MakeTLInboxMessageData(&inbox.InboxMessageData{
						RandomId:        outBox.RandomId,
						DialogMessageId: did,
						Message:         inboxMsg,
					}).To_InboxMessageData(),
				})
			if err != nil {
				c.Logger.Errorf("checkDuplicateMessage error - %v", err)
				return err
			}

			return err
		})
	if err != nil {
		c.Logger.Errorf("checkDuplicateMessage error - %v", err)
		return nil, err
	}

	return rUpdates, nil
}

func (c *MsgCore) sendChatMessage(
	ctx context.Context,
	fromUserId int64,
	fromAuthKeyId int64,
	chatId int64,
	outBox *msg.OutboxMessage,
	cb func(did int64, inboxMsg *mtproto.Message) error) (*mtproto.Updates, error) {

	hasDuplicateMessage, err := c.svcCtx.Dao.HasDuplicateMessage(ctx, fromUserId, outBox.RandomId)
	if err != nil {
		c.Logger.Errorf("checkDuplicateMessage error - %v", err)
		return nil, err
	} else if hasDuplicateMessage {
		rUpdates, err2 := c.svcCtx.Dao.GetDuplicateMessage(ctx, fromUserId, outBox.RandomId)
		if err2 != nil {
			c.Logger.Errorf("checkDuplicateMessage error - %v", err2)
			return nil, err2
		} else if rUpdates != nil {
			return rUpdates, nil
		}
	}

	box, ok, err := c.svcCtx.Dao.SendChatMessage(ctx, fromUserId, chatId, outBox)
	if err != nil {
		c.Logger.Error(err.Error())
		return nil, err
	}

	if ok && cb != nil {
		err = cb(box.DialogMessageId, box.ToMessage(fromUserId))
		if err != nil {
			c.Logger.Error(err.Error())
			return nil, err
		}
	}

	updateNewMessage := mtproto.MakeTLUpdateNewMessage(&mtproto.Update{
		Pts_INT32:       box.Pts,
		PtsCount:        box.PtsCount,
		RandomId:        box.RandomId,
		Message_MESSAGE: box.Message,
	}).To_Update()

	rUpdates := mtproto.MakeReplyUpdates(
		func(idList []int64) []*mtproto.User {
			users, _ := c.svcCtx.Dao.UserClient.UserGetMutableUsers(c.ctx,
				&userpb.TLUserGetMutableUsers{
					Id: idList,
				})
			return users.GetUserListByIdList(fromUserId, idList...)
		},
		func(idList []int64) []*mtproto.Chat {
			chats, _ := c.svcCtx.Dao.ChatClient.ChatGetChatListByIdList(c.ctx,
				&chatpb.TLChatGetChatListByIdList{
					IdList: idList,
				})
			return chats.GetChatListByIdList(fromUserId, idList...)
		},
		func(idList []int64) []*mtproto.Chat {
			// TODO
			return nil
		},
		updateNewMessage)

	if !ok {
		return rUpdates, nil
	}

	c.svcCtx.Dao.MessageDeDuplicate.PutDuplicateMessage(ctx, fromUserId, outBox.RandomId, rUpdates)

	c.svcCtx.Dao.SyncClient.SyncUpdatesNotMe(c.ctx, &sync.TLSyncUpdatesNotMe{
		UserId:        fromUserId,
		PermAuthKeyId: fromAuthKeyId,
		Updates: mtproto.MakeSyncNotMeUpdates(
			func(idList []int64) []*mtproto.User {
				return rUpdates.Users
			},
			func(idList []int64) []*mtproto.Chat {
				return rUpdates.Chats
			},
			func(idList []int64) []*mtproto.Chat {
				// rUpdates.Chats include chats
				return nil
			},
			updateNewMessage),
	})

	return rUpdates, nil
}

func (c *MsgCore) sendUserOutgoingMessageV2(fromUserId, fromAuthKeyId, toUserId int64, outBox *msg.OutboxMessage) (*mtproto.Updates, error) {
	var (
		idHelper = mtproto.NewIDListHelper(fromUserId, toUserId)
	)

	idHelper.PickByMessage(outBox.GetMessage())

	users, err := c.svcCtx.Dao.UserClient.UserGetMutableUsers(c.ctx, &userpb.TLUserGetMutableUsers{
		Id: idHelper.UserIdList,
		To: []int64{fromUserId, toUserId},
	})
	if err != nil {
		c.Logger.Errorf("msg.sendUserOutgoingMessageV2 - error: %v", err)
		return nil, err
	}

	sender, _ := users.GetImmutableUser(fromUserId)
	if sender == nil || sender.Deleted() {
		err = mtproto.ErrInputUserDeactivated
		c.Logger.Errorf("msg.sendUserOutgoingMessageV2 - error: %v", err)
		return nil, err
	}

	// TODO(@benqi): check
	// if sender.Restricted() {
	//	err = mtproto.ErrUserRestricted
	//	return
	// }

	peerUser, _ := users.GetImmutableUser(toUserId)
	if peerUser == nil || peerUser.Deleted() {
		err = mtproto.ErrInputUserDeactivated
		c.Logger.Errorf("msg.sendUserOutgoingMessage - error: %v", err)
		return nil, err
	}

	sendMe := fromUserId == toUserId
	if !sendMe {
		// TODO(@benqi)
		// 1. check blocked
		// 2. span
	}

	outBox.Message = plugin.RemakeMessage(
		c.ctx,
		c.svcCtx.MsgPlugin,
		outBox.Message,
		fromUserId,
		outBox.NoWebpage,
		func() bool {
			hasBot := false
			users.Visit(func(it *mtproto.ImmutableUser) {
				if it.IsBot() {
					hasBot = true
				}
			})

			return hasBot
		})

	var (
		// updateNewMessage *mtproto.Update
		rUpdates *mtproto.Updates
	)

	_, err = c.svcCtx.Dao.DoIdempotent(
		c.ctx,
		fromUserId,
		outBox.RandomId,
		&rUpdates,
		func(ctx context.Context, v any) error {
			box, err := c.svcCtx.Dao.SendUserMessageV2(ctx, fromUserId, toUserId, outBox)
			if err != nil {
				c.Logger.Error(err.Error())
				return err
			}

			_, err2 := c.svcCtx.Dao.InboxSendUserMessageToInboxV2(
				c.ctx,
				&inbox.TLInboxSendUserMessageToInboxV2{
					UserId:        fromUserId,
					Out:           true,
					FromId:        fromUserId,
					FromAuthKeyId: fromAuthKeyId,
					PeerType:      mtproto.PEER_USER,
					PeerId:        toUserId,
					Inbox:         box,
					Users:         users.GetUserListByIdList(fromUserId, idHelper.UserIdList...),
					Chats:         nil,
				})
			if err2 != nil {
				return err2
			}

			if fromUserId != toUserId {
				blocked, _ := c.svcCtx.Dao.UserClient.UserBlockedByUser(c.ctx, &userpb.TLUserBlockedByUser{
					UserId:     toUserId,
					PeerUserId: fromUserId,
				})

				if !mtproto.FromBool(blocked) {
					_, err2 = c.svcCtx.Dao.InboxSendUserMessageToInboxV2(
						c.ctx,
						&inbox.TLInboxSendUserMessageToInboxV2{
							UserId:        toUserId,
							Out:           false,
							FromId:        fromUserId,
							FromAuthKeyId: fromAuthKeyId,
							PeerType:      mtproto.PEER_USER,
							PeerId:        toUserId,
							Inbox:         box,
							Users:         users.GetUserListByIdList(toUserId, idHelper.UserIdList...),
							Chats:         nil,
						})
					if err2 != nil {
						return err2
					}
				}
			}

			*v.(**mtproto.Updates) = mtproto.MakeReplyUpdates(
				func(idList []int64) []*mtproto.User {
					// TODO: check
					//users, _ := c.svcCtx.Dao.UserClient.UserGetMutableUsers(ctx,
					//	&userpb.TLUserGetMutableUsers{
					//		Id: idList,
					//	})
					return users.GetUserListByIdList(fromUserId, idList...)
				},
				func(idList []int64) []*mtproto.Chat {
					return []*mtproto.Chat{}
				},
				func(idList []int64) []*mtproto.Chat {
					// TODO
					return []*mtproto.Chat{}
				},
				mtproto.MakeTLUpdateNewMessage(&mtproto.Update{
					Pts_INT32:       box.Pts,
					PtsCount:        box.PtsCount,
					RandomId:        box.RandomId,
					Message_MESSAGE: box.Message,
				}).To_Update())

			return nil
		})
	if err != nil {
		c.Logger.Errorf("msg.sendUserOutgoingMessageV2 - error: %v", err)
		return nil, err
	}

	return rUpdates, nil
}

func (c *MsgCore) sendChatOutgoingMessageV2(fromUserId, fromAuthKeyId, peerChatId int64, outBox *msg.OutboxMessage) (*mtproto.Updates, error) {
	var (
		chat      *mtproto.MutableChat
		sUserList *mtproto.MutableUsers
		idHelper  = mtproto.NewIDListHelper(fromUserId)
	)
	idHelper.PickByMessage(outBox.GetMessage())

	err := mr.Finish(
		func() error {
			var (
				err error
			)
			chat, err = c.svcCtx.Dao.ChatClient.ChatGetMutableChat(
				c.ctx,
				&chatpb.TLChatGetMutableChat{
					ChatId: peerChatId,
				})
			if err != nil {
				c.Logger.Errorf("inbox.sendChatMessageToInbox - error: %v", err)
			}
			return err
		},
		func() error {
			var (
				err error
			)
			sUserList, err = c.svcCtx.Dao.UserClient.UserGetMutableUsersV2(
				c.ctx,
				&userpb.TLUserGetMutableUsersV2{
					Id:      idHelper.UserIdList,
					Privacy: true,
					HasTo:   true,
					To:      nil,
				})
			if err != nil {
				c.Logger.Errorf("inbox.sendChatMessageToInbox - error: %v", err)
			}

			return err
		})
	if err != nil {
		// c.Logger.Errorf("inbox.sendChatMessageToInbox - error: %v", err)
		return nil, err
	}

	if _, ok := chat.GetImmutableChatParticipant(fromUserId); !ok {
		c.Logger.Errorf("msg.sendChatOutgoingMessageV2 - error: ErrChatParticipantNotExists")
		err = mtproto.ErrChatWriteForbidden
		return nil, err
	}

	outBox.Message = plugin.RemakeMessage(
		c.ctx,
		c.svcCtx.MsgPlugin,
		outBox.Message,
		fromUserId,
		outBox.NoWebpage,
		func() bool {
			hasBot := false
			chat.Walk(func(userId int64, participant *mtproto.ImmutableChatParticipant) error {
				if participant.IsBot {
					hasBot = true
				}
				return nil
			})

			return hasBot
		})

	var (
		rUpdates *mtproto.Updates
	)

	_, err = c.svcCtx.Dao.DoIdempotent(
		c.ctx,
		fromUserId,
		outBox.RandomId,
		&rUpdates,
		func(ctx context.Context, v any) error {
			box, err2 := c.svcCtx.Dao.SendChatMessageV2(ctx, fromUserId, peerChatId, outBox)
			if err2 != nil {
				c.Logger.Error(err2.Error())
				return err
			}

			chat.Walk(func(userId int64, participant *mtproto.ImmutableChatParticipant) error {
				if !participant.IsChatMemberStateNormal() {
					return nil
				}
				if err2 != nil {
					return nil
				}

				_, err2 = c.svcCtx.Dao.InboxClient.InboxSendUserMessageToInboxV2(ctx, &inbox.TLInboxSendUserMessageToInboxV2{
					UserId:        participant.UserId,
					Out:           participant.UserId == fromUserId,
					FromId:        fromUserId,
					FromAuthKeyId: fromAuthKeyId,
					PeerType:      mtproto.PEER_CHAT,
					PeerId:        peerChatId,
					Inbox:         box,
					Users:         sUserList.GetUserListByIdList(participant.UserId, idHelper.UserIdList...),
					Chats:         []*mtproto.Chat{chat.ToUnsafeChat(participant.UserId)},
				})
				return nil
			})

			if err2 != nil {
				c.Logger.Error(err2.Error())
				return err
			}

			*v.(**mtproto.Updates) = mtproto.MakeReplyUpdates(
				func(idList []int64) []*mtproto.User {
					return sUserList.GetUserListByIdList(fromUserId, idList...)
				},
				func(idList []int64) []*mtproto.Chat {
					return []*mtproto.Chat{chat.ToUnsafeChat(fromUserId)}
				},
				func(idList []int64) []*mtproto.Chat {
					// TODO
					return nil
				},
				mtproto.MakeTLUpdateNewMessage(&mtproto.Update{
					Pts_INT32:       box.Pts,
					PtsCount:        box.PtsCount,
					RandomId:        box.RandomId,
					Message_MESSAGE: box.Message,
				}).To_Update())

			return nil
		})
	if err != nil {
		c.Logger.Errorf("msg.sendChatOutgoingMessageV2 - error: %v", err)
		return nil, err
	}

	return rUpdates, nil
}
