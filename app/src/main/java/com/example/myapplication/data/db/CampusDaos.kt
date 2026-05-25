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

    @Query("SELECT * FROM users WHERE student_id = :studentId LIMIT 1")
    suspend fun getByStudentId(studentId: String): UserEntity?

    @Query("SELECT * FROM users WHERE student_id = :studentId AND password = :password LIMIT 1")
    suspend fun login(studentId: String, password: String): UserEntity?

    @Query("SELECT * FROM users WHERE id = :id LIMIT 1")
    suspend fun getById(id: Long): UserEntity?

    @Query("SELECT * FROM users WHERE id = :id LIMIT 1")
    fun observeUser(id: Long): Flow<UserEntity?>
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
}

@Dao
interface FavoriteDao {
    @Insert
    suspend fun insert(row: FavoriteEntity)

    @Query("DELETE FROM favorites WHERE user_id = :userId AND product_id = :productId")
    suspend fun delete(userId: Long, productId: Long)

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
}

@Dao
interface OrderDao {
    @Insert
    suspend fun insert(order: OrderEntity): Long

    @Update
    suspend fun update(order: OrderEntity)

    @Query("SELECT * FROM orders WHERE id = :id LIMIT 1")
    suspend fun getById(id: Long): OrderEntity?

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
        SELECT * FROM orders
        WHERE product_id = :productId
          AND (status = 0 OR status = 1)
        LIMIT 1
        """,
    )
    suspend fun findOpenOrderForProduct(productId: Long): OrderEntity?
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
               ) as lastMessagePreview
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
}
