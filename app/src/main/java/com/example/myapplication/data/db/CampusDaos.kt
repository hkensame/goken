package com.example.myapplication.data.db

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.Query
import androidx.room.Update
import kotlinx.coroutines.flow.Flow

@Dao
interface UserDao {
    @Insert
    suspend fun insert(user: UserEntity): Long

    @Update
    suspend fun update(user: UserEntity)

    @Query("SELECT * FROM users WHERE student_id = :studentId LIMIT 1")
    suspend fun getByStudentId(studentId: String): UserEntity?

    @Query("SELECT * FROM users WHERE student_id = :studentId AND password = :password LIMIT 1")
    suspend fun login(studentId: String, password: String): UserEntity?

    @Query("SELECT * FROM users WHERE id = :id LIMIT 1")
    suspend fun getById(id: Long): UserEntity?

    @Query("SELECT * FROM users WHERE id = :id LIMIT 1")
    fun observeUser(id: Long): Flow<UserEntity?>

    @Query("UPDATE users SET nickname = :nickname, updated_at = :updatedAt WHERE id = :id")
    suspend fun updateNickname(id: Long, nickname: String, updatedAt: Long)

    @Query("UPDATE users SET avatar_path = :avatarPath, updated_at = :updatedAt WHERE id = :id")
    suspend fun updateAvatar(id: Long, avatarPath: String?, updatedAt: Long)

    @Query("UPDATE users SET phone = :phone, updated_at = :updatedAt WHERE id = :id")
    suspend fun updatePhone(id: Long, phone: String?, updatedAt: Long)

    @Query("SELECT COUNT(*) FROM users")
    suspend fun count(): Int
}

@Dao
interface ProductDao {
    @Insert
    suspend fun insert(product: ProductEntity): Long

    @Update
    suspend fun update(product: ProductEntity)

    @Query(
        """
        SELECT p.id, p.seller_id as sellerId, p.title, p.price_cents as priceCents,
               p.category, p.image_path as imageLocalPath, p.status, u.nickname as sellerNickname
        FROM products p
        INNER JOIN users u ON p.seller_id = u.id
        WHERE p.status = :status
          AND (LENGTH(:q) = 0 OR p.title LIKE '%' || :q || '%' OR p.description LIKE '%' || :q || '%')
          AND (:categoryAll = 1 OR p.category = :category)
        ORDER BY p.created_at DESC
        """,
    )
    fun observeProducts(
        q: String,
        category: String,
        categoryAll: Int,
        status: String,
    ): Flow<List<ProductListItem>>

    @Query("SELECT * FROM products WHERE id = :id LIMIT 1")
    suspend fun getById(id: Long): ProductEntity?

    @Query("SELECT * FROM products WHERE id = :id LIMIT 1")
    fun observeById(id: Long): Flow<ProductEntity?>

    @Query(
        """
        SELECT p.id, p.seller_id as sellerId, p.title, p.price_cents as priceCents,
               p.category, p.image_path as imageLocalPath, p.status, u.nickname as sellerNickname
        FROM products p
        INNER JOIN users u ON p.seller_id = u.id
        WHERE p.seller_id = :sellerId
        ORDER BY p.created_at DESC
        """,
    )
    fun observeBySeller(sellerId: Long): Flow<List<ProductListItem>>

    @Query(
        """
        SELECT p.id, p.seller_id as sellerId, p.title, p.price_cents as priceCents,
               p.category, p.image_path as imageLocalPath, p.status, u.nickname as sellerNickname
        FROM products p
        INNER JOIN users u ON p.seller_id = u.id
        WHERE p.seller_id = :sellerId AND p.status = :status
        ORDER BY p.created_at DESC
        """,
    )
    fun observeBySellerAndStatus(sellerId: Long, status: String): Flow<List<ProductListItem>>

    @Query("DELETE FROM products WHERE id = :id")
    suspend fun delete(id: Long)

    @Query("SELECT COUNT(*) FROM products WHERE seller_id = :sellerId AND status = :status")
    suspend fun countBySellerAndStatus(sellerId: Long, status: String): Int

    @Query("SELECT COUNT(*) FROM products WHERE status = :status")
    suspend fun countByStatus(status: String): Int
}

@Dao
interface FavoriteDao {
    @Insert
    suspend fun insert(row: FavoriteEntity)

    @Query("DELETE FROM favorites WHERE user_id = :userId AND product_id = :productId")
    suspend fun delete(userId: Long, productId: Long)

    @Query("DELETE FROM favorites WHERE product_id = :productId")
    suspend fun deleteByProduct(productId: Long)

    @Query("SELECT * FROM favorites WHERE user_id = :userId AND product_id = :productId LIMIT 1")
    suspend fun find(userId: Long, productId: Long): FavoriteEntity?

    @Query("SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = :userId AND product_id = :productId)")
    fun observeFavorite(userId: Long, productId: Long): Flow<Boolean>

    @Query(
        """
        SELECT p.id, p.seller_id as sellerId, p.title, p.price_cents as priceCents,
               p.category, p.image_path as imageLocalPath, p.status, u.nickname as sellerNickname
        FROM favorites f
        INNER JOIN products p ON f.product_id = p.id
        INNER JOIN users u ON p.seller_id = u.id
        WHERE f.user_id = :userId
        ORDER BY f.created_at DESC
        """,
    )
    fun observeFavorites(userId: Long): Flow<List<ProductListItem>>

    @Query("SELECT COUNT(*) FROM favorites WHERE user_id = :userId")
    fun observeFavoriteCount(userId: Long): Flow<Int>
}

@Dao
interface OrderDao {
    @Insert
    suspend fun insert(order: OrderEntity): Long

    @Update
    suspend fun update(order: OrderEntity)

    @Query("SELECT * FROM orders WHERE id = :id LIMIT 1")
    suspend fun getById(id: Long): OrderEntity?

    @Query("SELECT * FROM orders WHERE id = :id LIMIT 1")
    fun observeById(id: Long): Flow<OrderEntity?>

    @Query(
        """
        SELECT o.id as orderId, o.status, o.created_at as createdAt, o.updated_at as updatedAt,
               o.product_id as productId, p.title as productTitle, p.price_cents as priceCents,
               o.buyer_id as buyerId, o.seller_id as sellerId,
               bu.nickname as buyerNickname, su.nickname as sellerNickname
        FROM orders o
        INNER JOIN products p ON o.product_id = p.id
        INNER JOIN users bu ON o.buyer_id = bu.id
        INNER JOIN users su ON o.seller_id = su.id
        WHERE o.buyer_id = :userId OR o.seller_id = :userId
        ORDER BY o.updated_at DESC
        """,
    )
    fun observeMine(userId: Long): Flow<List<OrderWithDetails>>

    @Query(
        """
        SELECT o.id as orderId, o.status, o.created_at as createdAt, o.updated_at as updatedAt,
               o.product_id as productId, p.title as productTitle, p.price_cents as priceCents,
               o.buyer_id as buyerId, o.seller_id as sellerId,
               bu.nickname as buyerNickname, su.nickname as sellerNickname
        FROM orders o
        INNER JOIN products p ON o.product_id = p.id
        INNER JOIN users bu ON o.buyer_id = bu.id
        INNER JOIN users su ON o.seller_id = su.id
        WHERE (o.buyer_id = :userId OR o.seller_id = :userId)
          AND o.status = :status
        ORDER BY o.updated_at DESC
        """,
    )
    fun observeMineByStatus(userId: Long, status: Int): Flow<List<OrderWithDetails>>

    @Query(
        """
        SELECT * FROM orders
        WHERE product_id = :productId
          AND (status = 0 OR status = 1)
        LIMIT 1
        """,
    )
    suspend fun findOpenOrderForProduct(productId: Long): OrderEntity?

    @Query("SELECT COUNT(*) FROM orders WHERE (buyer_id = :userId OR seller_id = :userId) AND status = :status")
    suspend fun countByUserAndStatus(userId: Long, status: Int): Int
}

@Dao
interface ConversationDao {
    @Insert
    suspend fun insert(c: ConversationEntity): Long

    @Update
    suspend fun update(c: ConversationEntity)

    @Query(
        """
        SELECT * FROM conversations
        WHERE product_id = :productId AND user_a_id = :a AND user_b_id = :b
        LIMIT 1
        """,
    )
    suspend fun find(productId: Long, a: Long, b: Long): ConversationEntity?

    @Query("SELECT * FROM conversations WHERE id = :id LIMIT 1")
    suspend fun getById(id: Long): ConversationEntity?

    @Query("SELECT * FROM conversations WHERE id = :id LIMIT 1")
    fun observeById(id: Long): Flow<ConversationEntity?>

    @Query(
        """
        SELECT c.id as conversationId, c.product_id as productId, p.title as productTitle,
               CASE WHEN c.user_a_id = :userId THEN c.user_b_id ELSE c.user_a_id END as peerId,
               u.nickname as peerNickname,
               c.last_message_at as lastMessageAt,
               (
                 SELECT m.content FROM messages m
                 WHERE m.conversation_id = c.id
                 ORDER BY m.created_at DESC LIMIT 1
               ) as lastMessagePreview,
               (
                 SELECT COUNT(*) FROM messages m
                 WHERE m.conversation_id = c.id AND m.sender_id != :userId AND m.is_read = 0
               ) as unreadCount
        FROM conversations c
        INNER JOIN products p ON c.product_id = p.id
        INNER JOIN users u ON u.id = CASE WHEN c.user_a_id = :userId THEN c.user_b_id ELSE c.user_a_id END
        WHERE c.user_a_id = :userId OR c.user_b_id = :userId
        ORDER BY c.last_message_at DESC
        """,
    )
    fun observeConversations(userId: Long): Flow<List<ConversationSummary>>
}

@Dao
interface MessageDao {
    @Insert
    suspend fun insert(m: MessageEntity): Long

    @Query("SELECT * FROM messages WHERE conversation_id = :cid ORDER BY created_at ASC")
    fun observeMessages(cid: Long): Flow<List<MessageEntity>>

    @Query("UPDATE messages SET is_read = 1 WHERE conversation_id = :conversationId AND sender_id != :userId AND is_read = 0")
    suspend fun markConversationRead(conversationId: Long, userId: Long)

    @Query(
        """
        SELECT COUNT(*) FROM messages m
        INNER JOIN conversations c ON m.conversation_id = c.id
        WHERE (c.user_a_id = :userId OR c.user_b_id = :userId)
          AND m.sender_id != :userId
          AND m.is_read = 0
        """,
    )
    fun observeUnreadCount(userId: Long): Flow<Int>

    @Query("SELECT COUNT(*) FROM messages WHERE conversation_id = :conversationId AND sender_id != :userId AND is_read = 0")
    fun observeUnreadCountByConversation(conversationId: Long, userId: Long): Flow<Int>
}

@Dao
interface ProductImageDao {
    @Insert
    suspend fun insert(image: ProductImageEntity): Long

    @Insert
    suspend fun insertAll(images: List<ProductImageEntity>): List<Long>

    @Query("SELECT * FROM product_images WHERE product_id = :productId ORDER BY sort_order ASC")
    fun observeByProduct(productId: Long): Flow<List<ProductImageEntity>>

    @Query("SELECT * FROM product_images WHERE product_id = :productId ORDER BY sort_order ASC")
    suspend fun getByProduct(productId: Long): List<ProductImageEntity>

    @Query("DELETE FROM product_images WHERE product_id = :productId")
    suspend fun deleteByProduct(productId: Long)

    @Query("DELETE FROM product_images WHERE id = :id")
    suspend fun delete(id: Long)
}

@Dao
interface NotificationDao {
    @Insert
    suspend fun insert(notification: NotificationEntity): Long

    @Query("SELECT * FROM notifications WHERE user_id = :userId ORDER BY created_at DESC")
    fun observeByUser(userId: Long): Flow<List<NotificationEntity>>

    @Query("SELECT COUNT(*) FROM notifications WHERE user_id = :userId AND is_read = 0")
    fun observeUnreadCount(userId: Long): Flow<Int>

    @Query("UPDATE notifications SET is_read = 1 WHERE id = :id")
    suspend fun markRead(id: Long)

    @Query("UPDATE notifications SET is_read = 1 WHERE user_id = :userId AND is_read = 0")
    suspend fun markAllRead(userId: Long)

    @Query("DELETE FROM notifications WHERE id = :id")
    suspend fun delete(id: Long)

    @Query("DELETE FROM notifications WHERE user_id = :userId")
    suspend fun deleteByUser(userId: Long)
}
