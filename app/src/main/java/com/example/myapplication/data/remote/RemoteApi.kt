package com.example.myapplication.data.remote

import com.example.myapplication.data.db.*

/**
 * 远程数据源接口 — 后期用 Retrofit + 后端 API 实现。
 *
 * 当前未实现，Repository 检测到 api == null 时自动走 Room 本地逻辑。
 * 搭建后端后，实现此接口并注入 Repository 即可切换为远程优先模式。
 */
interface RemoteApi {

    // ── Auth ──

    /** 注册，返回新用户 id；学号重复抛异常 */
    suspend fun register(studentId: String, password: String, nickname: String, createdAt: Long): Long

    /** 登录验证，匹配则返回 UserEntity */
    suspend fun login(studentId: String, password: String): UserEntity?

    suspend fun updateProfile(id: Long, nickname: String, phone: String?, updatedAt: Long)

    suspend fun updateAvatar(id: Long, avatarPath: String?, updatedAt: Long)

    // ── Product ──

    /** 发布商品，返回新商品 id */
    suspend fun publish(
        sellerId: Long, title: String, description: String,
        priceCents: Long, category: String, imagePath: String?,
        status: String, createdAt: Long,
    ): Long

    suspend fun updateProductStatus(id: Long, status: String, updatedAt: Long)

    // ── Favorite ──

    suspend fun addFavorite(userId: Long, productId: Long, createdAt: Long)

    suspend fun removeFavorite(userId: Long, productId: Long)

    // ── Order ──

    /** 创建订单，返回新订单 id */
    suspend fun createOrder(
        productId: Long, buyerId: Long, sellerId: Long,
        status: Int, createdAt: Long, updatedAt: Long,
    ): Long

    suspend fun updateOrderStatus(id: Long, status: Int, updatedAt: Long)

    // ── Chat ──

    /** 查找或创建会话，返回 conversationId */
    suspend fun findOrCreateConversation(
        productId: Long, userAId: Long, userBId: Long, lastMessageAt: Long,
    ): Long

    /** 发送消息，返回 messageId */
    suspend fun sendMessage(
        conversationId: Long, senderId: Long, content: String, createdAt: Long,
    ): Long

    suspend fun markConversationRead(conversationId: Long, userId: Long)

    suspend fun updateConversationLastMessage(conversationId: Long, lastMessageAt: Long)

    // ── Sync（从远程拉取数据刷新本地 Room 缓存） ──

    suspend fun fullSync(userId: Long)

    suspend fun syncUsers()

    suspend fun syncProducts()

    suspend fun syncFavorites(userId: Long)

    suspend fun syncOrders(userId: Long)

    suspend fun syncConversations(userId: Long)

    suspend fun syncMessages(conversationId: Long)

    suspend fun syncNotifications(userId: Long)

    suspend fun syncProductImages()
}
