// Copyright 2024 Teamgram Authors
//  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Author: teamgramio (teamgram.io@gmail.com)
//

package core

import (
	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/messenger/msg/inbox/inbox"
	"github.com/teamgram/teamgram-server/app/messenger/msg/internal/dal/dataobject"
	"github.com/teamgram/teamgram-server/app/messenger/sync/sync"
)

// InboxSendUserMessageToInboxV2
// inbox.sendUserMessageToInboxV2 flags:# user_id:long out:flags.0?true from_id:long peer_user_id:long inbox:MessageBox users:flags.1?Vector<ImmutableUser> = Void;
func (c *InboxCore) InboxSendUserMessageToInboxV2(in *inbox.TLInboxSendUserMessageToInboxV2) (*mtproto.Void, error) {
	if in.Out {
		err := c.svcCtx.Dao.SendMessageToOutboxV1(
			c.ctx,
			in.FromId,
			mtproto.MakeUserPeerUtil(in.PeerUserId),
			in.GetInbox())
		if err != nil {
			// TODO: handle error
			c.Logger.Errorf("inbox.sendUserMessageToInboxV2 - error: %v", err)
			return nil, err
		}

		// TODO: handle sendToSelfUser
		if in.FromId == in.PeerUserId {
			peer2 := in.GetInbox().GetMessage().GetSavedPeerId()
			if peer2 == nil {
				c.Logger.Errorf("inbox.sendUserMessageToInboxV2 - error: sendToSelfUser")
			} else {
				peer := mtproto.FromPeer(peer2)
				c.svcCtx.Dao.SavedDialogsDAO.InsertOrUpdate(
					c.ctx,
					&dataobject.SavedDialogsDO{
						UserId:     in.FromId,
						PeerType:   peer.PeerType,
						PeerId:     peer.PeerId,
						Pinned:     0,
						TopMessage: in.GetInbox().GetMessageId(),
					})
			}

			return mtproto.EmptyVoid, nil
		}
	} else {
		inBox, err := c.svcCtx.Dao.SendUserMessageToInbox(c.ctx,
			in.FromId,
			in.PeerUserId,
			in.GetInbox().GetDialogMessageId(),
			in.GetInbox().GetRandomId(),
			in.GetInbox().GetMessage())
		if err != nil {
			c.Logger.Errorf("inbox.sendUserMessageToInboxV2 - error: %v", err)
			return nil, err
		}

		if inBox.DialogMessageId == 1 &&
			(in.FromId != 42777 && in.FromId != 424000) {
			//isContact, _ := s.UserFacade.GetContactAndMutual(ctx, toId, fromId)
			//if !isContact {
			//	s.UserFacade.AddPeerSettings(ctx, toId, model.MakeUserPeerUtil(fromId), &mtproto.PeerSettings{
			//		AddContact:   true,
			//		BlockContact: true,
			//	})
			//}
		}

		var (
			pushUpdates = mtproto.MakeEmptyUpdates()
			inBoxHelper = mtproto.MakeBoxListByBoxListUsers([]*mtproto.MessageBox{inBox}, in.GetUsers())
		)

		inBoxHelper.Visit(
			inBox.UserId,
			func(messageList []*mtproto.Message) {
				for _, m := range messageList {
					pushUpdates.PushFrontUpdate(mtproto.MakeTLUpdateNewMessage(&mtproto.Update{
						Message_MESSAGE: m,
						Pts_INT32:       inBox.Pts,
						PtsCount:        inBox.PtsCount,
					}).To_Update())
				}
			},
			func(users []*mtproto.User, rawIdList []int64) {
				pushUpdates.PushUser(users...)
			},
			func(chats []*mtproto.Chat, rawIdList []int64) {
				pushUpdates.PushChat(chats...)
			},
			func(chats []*mtproto.Chat, rawIdList []int64) {
				pushUpdates.PushChat(chats...)
			})

		var (
			isBot = false
		)

		for _, u := range pushUpdates.GetUsers() {
			if u.GetId() == in.PeerUserId {
				isBot = u.GetBot()
				break
			}
		}

		if isBot {
			if c.svcCtx.Dao.BotSyncClient != nil {
				_, err = c.svcCtx.Dao.BotSyncClient.SyncPushBotUpdates(c.ctx, &sync.TLSyncPushBotUpdates{
					UserId:  inBox.UserId,
					Updates: pushUpdates,
				})
			} else {
				// TODO: log
			}
		} else {
			_, err = c.svcCtx.Dao.SyncClient.SyncPushUpdates(c.ctx, &sync.TLSyncPushUpdates{
				UserId:  inBox.UserId,
				Updates: pushUpdates,
			})
		}
		if err != nil {
			c.Logger.Errorf("inbox.sendUserMessageToInboxV2 - error: %v", err)
		}
	}

	return mtproto.EmptyVoid, nil
}