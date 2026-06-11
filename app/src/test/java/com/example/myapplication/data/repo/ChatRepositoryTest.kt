package com.example.myapplication.data.repo

import com.example.myapplication.data.db.CampusDatabase
import com.example.myapplication.data.db.ConversationDao
import com.example.myapplication.data.db.ConversationEntity
import com.example.myapplication.data.db.MessageDao
import com.example.myapplication.data.db.MessageEntity
import com.example.myapplication.data.remote.MockRemoteApi
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import io.mockk.slot
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class ChatRepositoryTest {

    private val db = mockk<CampusDatabase>(relaxed = true)
    private val conversationDao = mockk<ConversationDao>(relaxed = true)
    private val messageDao = mockk<MessageDao>(relaxed = true)

    init {
        coEvery { db.conversationDao() } returns conversationDao
        coEvery { db.messageDao() } returns messageDao
    }

    // ════════════════════════════════════════
    //  Room 分支（api == null）— 无事务依赖
    // ════════════════════════════════════════

    @Test
    fun room_sendMessage_emptyContent_ignored() = runTest {
        val repo = ChatRepository(db)
        repo.sendMessage(1L, 1L, "   ")
        coVerify(exactly = 0) { messageDao.insert(any()) }
    }

    @Test
    fun room_markAsRead_callsDao() = runTest {
        val repo = ChatRepository(db)
        repo.markAsRead(5L, 1L)
        coVerify { messageDao.markConversationRead(5L, 1L) }
    }

    // ════════════════════════════════════════
    //  MockRemoteApi 分支 — 核心测试
    // ════════════════════════════════════════

    @Test
    fun api_getOrCreateConversation_callsMockAndReturnsId() = runTest {
        val api = MockRemoteApi()
        val repo = ChatRepository(db, api)

        val id = repo.getOrCreateConversation(10L, selfId = 1L, peerId = 2L)
        assertEquals(1L, id)
        assertTrue(api.called("findOrCreateConversation"))
        assertTrue(api.called("syncConversations"))
        // mock 成功不走 Room
        coVerify(exactly = 0) { conversationDao.find(any(), any(), any()) }
    }

    @Test
    fun api_getOrCreateConversation_swapsUserIds() = runTest {
        val api = MockRemoteApi()
        val repo = ChatRepository(db, api)

        // selfId=3, peerId=1 → 内部 min(3,1)=1, max(3,1)=3
        repo.getOrCreateConversation(10L, selfId = 3L, peerId = 1L)
        val conv = api.getConversation(1L)
        assertEquals(1L, conv!!.userAId)
        assertEquals(3L, conv.userBId)
    }

    @Test
    fun api_sendMessage_callsMockAndSyncs() = runTest {
        val api = MockRemoteApi()
        val repo = ChatRepository(db, api)

        repo.sendMessage(1L, 1L, "你好")
        assertTrue(api.called("sendMessage"))
        assertTrue(api.called("updateConversationLastMessage"))
        assertTrue(api.called("syncMessages"))
        // mock 成功不走 Room
        coVerify(exactly = 0) { messageDao.insert(any()) }
    }

    @Test
    fun api_sendMessage_emptyContent_stillIgnored() = runTest {
        val api = MockRemoteApi()
        val repo = ChatRepository(db, api)

        repo.sendMessage(1L, 1L, "   ")
        assertEquals(0, api.calls.size)
    }

    @Test
    fun api_sendMessage_trimsContent() = runTest {
        val api = MockRemoteApi()
        api.findOrCreateConversation(10L, 1L, 2L, 1000L)
        val repo = ChatRepository(db, api)

        repo.sendMessage(1L, 1L, "  你好  ")
        val msgs = api.getMessages(1L)
        assertEquals("你好", msgs[0].content)
    }

    @Test
    fun api_markAsRead_callsMockAndRoom() = runTest {
        val api = MockRemoteApi()
        val repo = ChatRepository(db, api)

        repo.markAsRead(5L, 1L)
        assertTrue(api.called("markConversationRead"))
        coVerify { messageDao.markConversationRead(5L, 1L) }
    }

    @Test
    fun api_getOrCreateConversation_existing_returnsSameId() = runTest {
        val api = MockRemoteApi()
        val repo = ChatRepository(db, api)

        val id1 = repo.getOrCreateConversation(10L, selfId = 1L, peerId = 2L)
        val id2 = repo.getOrCreateConversation(10L, selfId = 1L, peerId = 2L)
        assertEquals(id1, id2) // 幂等
    }
}