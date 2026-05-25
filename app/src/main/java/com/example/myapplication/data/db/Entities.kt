package com.example.myapplication.data.db

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.ForeignKey
import androidx.room.Index
import androidx.room.PrimaryKey

@Entity(
    tableName = "users",
    indices = [Index(value = ["student_id"], unique = true)],
)
data class UserEntity(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    @ColumnInfo(name = "student_id") val studentId: String,
    val password: String,
    val nickname: String,
    val createdAt: Long,
)

@Entity(
    tableName = "products",
    foreignKeys = [
        ForeignKey(
            entity = UserEntity::class,
            parentColumns = ["id"],
            childColumns = ["seller_id"],
            onDelete = ForeignKey.CASCADE,
        ),
    ],
    indices = [Index("seller_id"), Index("category"), Index("status")],
)
data class ProductEntity(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    @ColumnInfo(name = "seller_id") val sellerId: Long,
    val title: String,
    val description: String,
    @ColumnInfo(name = "price_cents") val priceCents: Long,
    val category: String,
    @ColumnInfo(name = "image_path") val imageLocalPath: String?,
    val status: String,
    @ColumnInfo(name = "created_at") val createdAt: Long,
)

@Entity(
    tableName = "favorites",
    primaryKeys = ["user_id", "product_id"],
    foreignKeys = [
        ForeignKey(
            entity = UserEntity::class,
            parentColumns = ["id"],
            childColumns = ["user_id"],
            onDelete = ForeignKey.CASCADE,
        ),
        ForeignKey(
            entity = ProductEntity::class,
            parentColumns = ["id"],
            childColumns = ["product_id"],
            onDelete = ForeignKey.CASCADE,
        ),
    ],
    indices = [Index("user_id"), Index("product_id")],
)
data class FavoriteEntity(
    @ColumnInfo(name = "user_id") val userId: Long,
    @ColumnInfo(name = "product_id") val productId: Long,
    @ColumnInfo(name = "created_at") val createdAt: Long,
)

@Entity(
    tableName = "orders",
    foreignKeys = [
        ForeignKey(entity = ProductEntity::class, parentColumns = ["id"], childColumns = ["product_id"], onDelete = ForeignKey.CASCADE),
        ForeignKey(entity = UserEntity::class, parentColumns = ["id"], childColumns = ["buyer_id"], onDelete = ForeignKey.CASCADE),
        ForeignKey(entity = UserEntity::class, parentColumns = ["id"], childColumns = ["seller_id"], onDelete = ForeignKey.CASCADE),
    ],
    indices = [Index("buyer_id"), Index("seller_id"), Index("product_id"), Index("status")],
)
data class OrderEntity(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    @ColumnInfo(name = "product_id") val productId: Long,
    @ColumnInfo(name = "buyer_id") val buyerId: Long,
    @ColumnInfo(name = "seller_id") val sellerId: Long,
    val status: Int,
    @ColumnInfo(name = "created_at") val createdAt: Long,
    @ColumnInfo(name = "updated_at") val updatedAt: Long,
)

@Entity(
    tableName = "conversations",
    foreignKeys = [
        ForeignKey(entity = ProductEntity::class, parentColumns = ["id"], childColumns = ["product_id"], onDelete = ForeignKey.CASCADE),
        ForeignKey(entity = UserEntity::class, parentColumns = ["id"], childColumns = ["user_a_id"], onDelete = ForeignKey.CASCADE),
        ForeignKey(entity = UserEntity::class, parentColumns = ["id"], childColumns = ["user_b_id"], onDelete = ForeignKey.CASCADE),
    ],
    indices = [
        Index(value = ["product_id", "user_a_id", "user_b_id"], unique = true),
    ],
)
data class ConversationEntity(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    @ColumnInfo(name = "product_id") val productId: Long,
    @ColumnInfo(name = "user_a_id") val userAId: Long,
    @ColumnInfo(name = "user_b_id") val userBId: Long,
    @ColumnInfo(name = "last_message_at") val lastMessageAt: Long,
)

@Entity(
    tableName = "messages",
    foreignKeys = [
        ForeignKey(
            entity = ConversationEntity::class,
            parentColumns = ["id"],
            childColumns = ["conversation_id"],
            onDelete = ForeignKey.CASCADE,
        ),
        ForeignKey(
            entity = UserEntity::class,
            parentColumns = ["id"],
            childColumns = ["sender_id"],
            onDelete = ForeignKey.CASCADE,
        ),
    ],
    indices = [Index("conversation_id"), Index("created_at")],
)
data class MessageEntity(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    @ColumnInfo(name = "conversation_id") val conversationId: Long,
    @ColumnInfo(name = "sender_id") val senderId: Long,
    val content: String,
    @ColumnInfo(name = "created_at") val createdAt: Long,
)

/** 列表展示用：商品 + 卖家昵称 */
data class ProductListItem(
    val id: Long,
    val sellerId: Long,
    val title: String,
    val priceCents: Long,
    val category: String,
    val imageLocalPath: String?,
    val status: String,
    val sellerNickname: String,
)

data class OrderWithDetails(
    val orderId: Long,
    val status: Int,
    val createdAt: Long,
    val updatedAt: Long,
    val productId: Long,
    val productTitle: String,
    val priceCents: Long,
    val buyerId: Long,
    val sellerId: Long,
    val buyerNickname: String,
    val sellerNickname: String,
)

data class ConversationSummary(
    val conversationId: Long,
    val productId: Long,
    val productTitle: String,
    val peerId: Long,
    val peerNickname: String,
    val lastMessageAt: Long,
    val lastMessagePreview: String?,
)
