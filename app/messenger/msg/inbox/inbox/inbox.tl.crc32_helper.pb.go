/*
 * WARNING! All changes made in this file will be lost!
 * Created from 'scheme.tl' by 'mtprotoc'
 *
 * Copyright (c) 2024-present,  Teamgram Authors.
 *  All rights reserved.
 *
 * Author: Benqi (wubenqi@gmail.com)
 */

package inbox

const (
	CRC32_UNKNOWN                           TLConstructor = 0
	CRC32_inboxMessageData                  TLConstructor = 1002286548  // 0x3bbdadd4
	CRC32_inboxMessageId                    TLConstructor = -963460705  // 0xc692c19f
	CRC32_inbox_sendUserMessageToInbox      TLConstructor = -208741709  // 0xf38edab3
	CRC32_inbox_sendChatMessageToInbox      TLConstructor = -1760197438 // 0x971584c2
	CRC32_inbox_sendUserMultiMessageToInbox TLConstructor = -1782288007 // 0x95c47179
	CRC32_inbox_sendChatMultiMessageToInbox TLConstructor = -694455924  // 0xd69b718c
	CRC32_inbox_editUserMessageToInbox      TLConstructor = 1559967656  // 0x5cfb37a8
	CRC32_inbox_editChatMessageToInbox      TLConstructor = 2031122959  // 0x79107a0f
	CRC32_inbox_deleteMessagesToInbox       TLConstructor = -2061734348 // 0x851c6e34
	CRC32_inbox_deleteUserHistoryToInbox    TLConstructor = 336232792   // 0x140a8158
	CRC32_inbox_deleteChatHistoryToInbox    TLConstructor = -659905022  // 0xd8aaa602
	CRC32_inbox_readUserMediaUnreadToInbox  TLConstructor = 364970827   // 0x15c1034b
	CRC32_inbox_readChatMediaUnreadToInbox  TLConstructor = 1430347220  // 0x55415dd4
	CRC32_inbox_updateHistoryReaded         TLConstructor = -1010283296 // 0xc3c84ce0
	CRC32_inbox_updatePinnedMessage         TLConstructor = -1452528908 // 0xa96c2af4
	CRC32_inbox_unpinAllMessages            TLConstructor = 589079137   // 0x231ca261
)