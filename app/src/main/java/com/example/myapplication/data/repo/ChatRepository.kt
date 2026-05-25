package com.example.myapplication.data.repo

import androidx.room.withTransaction
import com.example.myapplication.data.db.CampusDatabase
import com.example.myapplication.data.db.ConversationEntity
import com.example.myapplication.data.db.MessageEntity
import kotlinx.coroutines.flow.Flow

class ChatRepository(
    private val db: CampusDatabase,
) {
    private val c = db.conversationDao()
    private val m = db.messageDao()

    fun observeConversations(userId: Long): Flow<List<com.example.myapplication.data.db.ConversationSummary>> =
        c.observeConversations(userId)

    fun observeMessages(conversationId: Long) = m.observeMessages(conversationId)

    suspend fun getOrCreateConversation(productId: Long, selfId: Long, peerId: Long): Long =
        db.withTransaction {
            val a = minOf(selfId, peerId)
            val b = maxOf(selfId, peerId)
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

    suspend fun sendMessage(conversationId: Long, senderId: Long, content: String) {
        val text = content.trim()
        if (text.isEmpty()) return
        val now = System.currentTimeMillis()
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
