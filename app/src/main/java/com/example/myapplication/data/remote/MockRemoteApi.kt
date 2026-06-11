package com.example.myapplication.data.remote

import com.example.myapplication.data.db.*
import kotlinx.coroutines.delay

/**
 * RemoteApi 的内存 Mock 实现。
 *
 * 用途：
 * 1. 单元测试 — 不依赖 Room / 网络，直接验证 Repository 的远程分支逻辑
 * 2. 开发模式 — 注入 container.api，无需后端即可跑通完整流程
 * 3. 集成测试 — 传入 roomDb 后 sync 方法会写入 Room，模拟真实同步
 *
 * 示例：
 *   val mock = MockRemoteApi()
 *   val id = mock.register("S1", "pw", "nick", now)  // 返回 1
 *   mock.login("S1", "pw")  // 返回 UserEntity
 */
class MockRemoteApi(
    /** 传入 Room 数据库后，sync 方法会将内存数据写入 Room；传 null 则 sync 为空操作 */
    private val roomDb: CampusDatabase? = null,
    /** 模拟网络延迟（毫秒），0 = 不延迟 */
    private val networkDelayMs: Long = 0,
) : RemoteApi {

    // ── 内存存储 ──

    private val users = mutableMapOf<Long, UserEntity>()
    private val products = mutableMapOf<Long, ProductEntity>()
    private val favorites = mutableListOf<FavoriteEntity>()
    private val orders = mutableMapOf<Long, OrderEntity>()
    private val conversations = mutableMapOf<Long, ConversationEntity>()
    private val messages = mutableMapOf<Long, MessageEntity>()
    private val notifications = mutableMapOf<Long, NotificationEntity>()
    private val productImages = mutableMapOf<Long, ProductImageEntity>()

    // ── 自增 ID ──

    private var nextUserId = 1L
    private var nextProductId = 1L
    private var nextOrderId = 1L
    private var nextConversationId = 1L
    private var nextMessageId = 1L
    private var nextNotificationId = 1L
    private var nextProductImageId = 1L

    // ── 调用记录（方便测试断言） ──

    val calls = mutableListOf<String>()

    private suspend fun simulateNetwork() {
        if (networkDelayMs > 0) delay(networkDelayMs)
    }

    // ══════════════════════════════════════════════════
    //  Auth
    // ══════════════════════════════════════════════════

    override suspend fun register(studentId: String, password: String, nickname: String, createdAt: Long): Long {
        simulateNetwork()
        calls += "register($studentId)"
        if (users.values.any { it.studentId == studentId }) {
            throw IllegalStateException("该学号已注册")
        }
        val id = nextUserId++
        users[id] = UserEntity(
            id = id, studentId = studentId, password = password,
            nickname = nickname, createdAt = createdAt,
        )
        return id
    }

    override suspend fun login(studentId: String, password: String): UserEntity? {
        simulateNetwork()
        calls += "login($studentId)"
        return users.values.find { it.studentId == studentId && it.password == password }
    }

    override suspend fun updateProfile(id: Long, nickname: String, phone: String?, updatedAt: Long) {
        simulateNetwork()
        calls += "updateProfile($id)"
        users[id]?.let { existing ->
            users[id] = existing.copy(nickname = nickname, phone = phone, updatedAt = updatedAt)
        }
    }

    override suspend fun updateAvatar(id: Long, avatarPath: String?, updatedAt: Long) {
        simulateNetwork()
        calls += "updateAvatar($id)"
        users[id]?.let { existing ->
            users[id] = existing.copy(avatarPath = avatarPath, updatedAt = updatedAt)
        }
    }

    // ══════════════════════════════════════════════════
    //  Product
    // ══════════════════════════════════════════════════

    override suspend fun publish(
        sellerId: Long, title: String, description: String,
        priceCents: Long, category: String, imagePath: String?,
        status: String, createdAt: Long,
    ): Long {
        simulateNetwork()
        calls += "publish($sellerId,$title)"
        val id = nextProductId++
        products[id] = ProductEntity(
            id = id, sellerId = sellerId, title = title,
            description = description, priceCents = priceCents,
            category = category, imageLocalPath = imagePath,
            status = status, createdAt = createdAt,
        )
        return id
    }

    override suspend fun updateProductStatus(id: Long, status: String, updatedAt: Long) {
        simulateNetwork()
        calls += "updateProductStatus($id,$status)"
        products[id]?.let { existing ->
            products[id] = existing.copy(status = status, updatedAt = updatedAt)
        }
    }

    // ══════════════════════════════════════════════════
    //  Favorite
    // ══════════════════════════════════════════════════

    override suspend fun addFavorite(userId: Long, productId: Long, createdAt: Long) {
        simulateNetwork()
        calls += "addFavorite($userId,$productId)"
        val entity = FavoriteEntity(userId = userId, productId = productId, createdAt = createdAt)
        if (favorites.none { it.userId == userId && it.productId == productId }) {
            favorites += entity
        }
    }

    override suspend fun removeFavorite(userId: Long, productId: Long) {
        simulateNetwork()
        calls += "removeFavorite($userId,$productId)"
        favorites.removeAll { it.userId == userId && it.productId == productId }
    }

    // ══════════════════════════════════════════════════
    //  Order
    // ══════════════════════════════════════════════════

    override suspend fun createOrder(
        productId: Long, buyerId: Long, sellerId: Long,
        status: Int, createdAt: Long, updatedAt: Long,
    ): Long {
        simulateNetwork()
        calls += "createOrder($productId,$buyerId)"
        val id = nextOrderId++
        orders[id] = OrderEntity(
            id = id, productId = productId, buyerId = buyerId,
            sellerId = sellerId, status = status,
            createdAt = createdAt, updatedAt = updatedAt,
        )
        return id
    }

    override suspend fun updateOrderStatus(id: Long, status: Int, updatedAt: Long) {
        simulateNetwork()
        calls += "updateOrderStatus($id,$status)"
        orders[id]?.let { existing ->
            orders[id] = existing.copy(status = status, updatedAt = updatedAt)
        }
    }

    // ══════════════════════════════════════════════════
    //  Chat
    // ══════════════════════════════════════════════════

    override suspend fun findOrCreateConversation(
        productId: Long, userAId: Long, userBId: Long, lastMessageAt: Long,
    ): Long {
        simulateNetwork()
        calls += "findOrCreateConversation($productId,$userAId,$userBId)"
        val existing = conversations.values.find {
            it.productId == productId && it.userAId == userAId && it.userBId == userBId
        }
        if (existing != null) return existing.id
        val id = nextConversationId++
        conversations[id] = ConversationEntity(
            id = id, productId = productId,
            userAId = userAId, userBId = userBId,
            lastMessageAt = lastMessageAt,
        )
        return id
    }

    override suspend fun sendMessage(
        conversationId: Long, senderId: Long, content: String, createdAt: Long,
    ): Long {
        simulateNetwork()
        calls += "sendMessage($conversationId,$senderId)"
        val id = nextMessageId++
        messages[id] = MessageEntity(
            id = id, conversationId = conversationId,
            senderId = senderId, content = content,
            createdAt = createdAt,
        )
        return id
    }

    override suspend fun markConversationRead(conversationId: Long, userId: Long) {
        simulateNetwork()
        calls += "markConversationRead($conversationId,$userId)"
        messages.values.forEach { msg ->
            if (msg.conversationId == conversationId && msg.senderId != userId && msg.isRead == 0) {
                messages[msg.id] = msg.copy(isRead = 1)
            }
        }
    }

    override suspend fun updateConversationLastMessage(conversationId: Long, lastMessageAt: Long) {
        simulateNetwork()
        calls += "updateConversationLastMessage($conversationId)"
        conversations[conversationId]?.let { existing ->
            conversations[conversationId] = existing.copy(lastMessageAt = lastMessageAt)
        }
    }

    // ══════════════════════════════════════════════════
    //  Sync（内存 → Room）
    // ══════════════════════════════════════════════════

    override suspend fun fullSync(userId: Long) {
        syncUsers()
        syncProducts()
        syncFavorites(userId)
        syncOrders(userId)
        syncConversations(userId)
        syncNotifications(userId)
        syncProductImages()
    }

    override suspend fun syncUsers() {
        calls += "syncUsers"
        val db = roomDb ?: return
        users.values.forEach { entity ->
            val existing = db.userDao().getById(entity.id)
            if (existing == null) db.userDao().insert(entity)
            else db.userDao().update(entity)
        }
    }

    override suspend fun syncProducts() {
        calls += "syncProducts"
        val db = roomDb ?: return
        products.values.forEach { entity ->
            val existing = db.productDao().getById(entity.id)
            if (existing == null) db.productDao().insert(entity)
            else db.productDao().update(entity)
        }
    }

    override suspend fun syncFavorites(userId: Long) {
        calls += "syncFavorites($userId)"
        val db = roomDb ?: return
        favorites.filter { it.userId == userId }.forEach { entity ->
            val existing = db.favoriteDao().find(entity.userId, entity.productId)
            if (existing == null) db.favoriteDao().insert(entity)
        }
    }

    override suspend fun syncOrders(userId: Long) {
        calls += "syncOrders($userId)"
        val db = roomDb ?: return
        orders.values.filter { it.buyerId == userId || it.sellerId == userId }.forEach { entity ->
            val existing = db.orderDao().getById(entity.id)
            if (existing == null) db.orderDao().insert(entity)
            else db.orderDao().update(entity)
        }
    }

    override suspend fun syncConversations(userId: Long) {
        calls += "syncConversations($userId)"
        val db = roomDb ?: return
        conversations.values.filter { it.userAId == userId || it.userBId == userId }.forEach { entity ->
            val existing = db.conversationDao().getById(entity.id)
            if (existing == null) db.conversationDao().insert(entity)
            else db.conversationDao().update(entity)
        }
        // 同步这些会话的消息
        conversations.values.filter { it.userAId == userId || it.userBId == userId }.forEach { conv ->
            syncMessages(conv.id)
        }
    }

    override suspend fun syncMessages(conversationId: Long) {
        calls += "syncMessages($conversationId)"
        val db = roomDb ?: return
        messages.values.filter { it.conversationId == conversationId }.forEach { entity ->
            val existing = db.messageDao().getById(entity.id)
            if (existing == null) db.messageDao().insert(entity)
        }
    }

    override suspend fun syncNotifications(userId: Long) {
        calls += "syncNotifications($userId)"
        val db = roomDb ?: return
        notifications.values.filter { it.userId == userId }.forEach { entity ->
            val existing = db.notificationDao().getById(entity.id)
            if (existing == null) db.notificationDao().insert(entity)
        }
    }

    override suspend fun syncProductImages() {
        calls += "syncProductImages"
        val db = roomDb ?: return
        productImages.values.forEach { entity ->
            val existing = db.productImageDao().getById(entity.id)
            if (existing == null) db.productImageDao().insert(entity)
        }
    }

    // ══════════════════════════════════════════════════
    //  测试辅助方法
    // ══════════════════════════════════════════════════

    /** 清空所有内存数据和调用记录 */
    fun clear() {
        users.clear(); products.clear(); favorites.clear()
        orders.clear(); conversations.clear(); messages.clear()
        notifications.clear(); productImages.clear()
        nextUserId = 1; nextProductId = 1; nextOrderId = 1
        nextConversationId = 1; nextMessageId = 1
        nextNotificationId = 1; nextProductImageId = 1
        calls.clear()
    }

    /** 直接往内存灌入测试数据 */
    fun putUser(entity: UserEntity) { users[entity.id] = entity; nextUserId = maxOf(nextUserId, entity.id + 1) }
    fun putProduct(entity: ProductEntity) { products[entity.id] = entity; nextProductId = maxOf(nextProductId, entity.id + 1) }
    fun putFavorite(entity: FavoriteEntity) { favorites += entity }
    fun putOrder(entity: OrderEntity) { orders[entity.id] = entity; nextOrderId = maxOf(nextOrderId, entity.id + 1) }
    fun putConversation(entity: ConversationEntity) { conversations[entity.id] = entity; nextConversationId = maxOf(nextConversationId, entity.id + 1) }
    fun putMessage(entity: MessageEntity) { messages[entity.id] = entity; nextMessageId = maxOf(nextMessageId, entity.id + 1) }
    fun putNotification(entity: NotificationEntity) { notifications[entity.id] = entity; nextNotificationId = maxOf(nextNotificationId, entity.id + 1) }
    fun putProductImage(entity: ProductImageEntity) { productImages[entity.id] = entity; nextProductImageId = maxOf(nextProductImageId, entity.id + 1) }

    /** 读取内存数据（断言用） */
    fun getUser(id: Long): UserEntity? = users[id]
    fun getProduct(id: Long): ProductEntity? = products[id]
    fun getOrder(id: Long): OrderEntity? = orders[id]
    fun getConversation(id: Long): ConversationEntity? = conversations[id]
    fun getMessages(conversationId: Long): List<MessageEntity> = messages.values.filter { it.conversationId == conversationId }
    fun getFavorites(userId: Long): List<FavoriteEntity> = favorites.filter { it.userId == userId }

    /** 判断某个方法是否被调用过 */
    fun called(method: String): Boolean = calls.any { it.startsWith(method) }

    /** 调用次数 */
    fun callCount(method: String): Int = calls.count { it.startsWith(method) }
}
