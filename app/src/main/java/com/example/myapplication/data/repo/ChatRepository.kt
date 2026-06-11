package com.example.myapplication.data.repo

import androidx.room.withTransaction
import com.example.myapplication.data.db.CampusDatabase
import com.example.myapplication.data.db.ConversationEntity
import com.example.myapplication.data.db.MessageEntity
import com.example.myapplication.data.remote.RemoteApi
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.withContext

class ChatRepository(
    private val db: CampusDatabase,
    private val api: RemoteApi? = null,
) {
    constructor(db: CampusDatabase) : this(db, null)

    private val c = db.conversationDao()
    private val m = db.messageDao()

    fun observeConversations(userId: Long): Flow<List<com.example.myapplication.data.db.ConversationSummary>> =
        c.observeConversations(userId)

    fun observeMessages(conversationId: Long) = m.observeMessages(conversationId)

    fun observeUnreadCount(userId: Long) = m.observeUnreadCount(userId)

    suspend fun markAsRead(conversationId: Long, userId: Long) {
        try { api?.markConversationRead(conversationId, userId) } catch (_: Exception) {}
        m.markConversationRead(conversationId, userId)
        try { api?.syncMessages(conversationId) } catch (_: Exception) {}
    }

    suspend fun getOrCreateConversation(productId: Long, selfId: Long, peerId: Long): Long {
        val a = minOf(selfId, peerId)
        val b = maxOf(selfId, peerId)

        if (api != null) {
            try {
                return withContext(Dispatchers.IO) {
                    val now = System.currentTimeMillis()
                    val id = api.findOrCreateConversation(productId, a, b, now)
                    api.syncConversations(selfId)
                    id
                }
            } catch (_: Exception) { /* 远程失败，回退 Room */ }
        }

        return db.withTransaction {
            val existing = c.find(productId, a, b)
            if (existing != null) return@withTransaction existing.id
            val now = System.currentTimeMillis()
            c.insert(
                ConversationEntity(
                    productId = productId,
                    userAId = a,
                    userBId = b,
                    lastMessageAt = now,
                ),
            )
        }
    }

    suspend fun sendMessage(conversationId: Long, senderId: Long, content: String) {
        val text = content.trim()
        if (text.isEmpty()) return
        val now = System.currentTimeMillis()

        if (api != null) {
            try {
                withContext(Dispatchers.IO) {
                    api.sendMessage(conversationId, senderId, text, now)
                    api.updateConversationLastMessage(conversationId, now)
                    api.syncMessages(conversationId)
                    api.syncConversations(senderId)
                }
                return
            } catch (_: Exception) { /* 远程失败，回退 Room */ }
        }

        db.withTransaction {
            m.insert(
                MessageEntity(
                    conversationId = conversationId,
                    senderId = senderId,
                    content = text,
                    createdAt = now,
                ),
            )
            val conv = c.getById(conversationId) ?: return@withTransaction
            c.update(conv.copy(lastMessageAt = now))
        }
    }
}
