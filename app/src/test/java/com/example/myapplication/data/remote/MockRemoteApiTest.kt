package com.example.myapplication.data.remote

import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

/**
 * 纯内存测试 [MockRemoteApi] 的全部接口行为。
 * 不依赖 Room / 网络 / Android。
 */
class MockRemoteApiTest {

    private lateinit var api: MockRemoteApi

    @Before
    fun setup() { api = MockRemoteApi() }

    // ── Auth ──

    @Test
    fun register_returnsIncrementalId() = runTest {
        assertEquals(1L, api.register("S001", "pw1", "Alice", 1000L))
        assertEquals(2L, api.register("S002", "pw2", "Bob", 1001L))
    }

    @Test
    fun register_duplicate_throws() = runTest {
        api.register("S001", "pw", "A", 1000L)
        try {
            api.register("S001", "pw2", "B", 1001L)
            fail("应该抛异常")
        } catch (_: IllegalStateException) {}
    }

    @Test
    fun login_correctPassword_returnsUser() = runTest {
        api.register("S001", "pw", "Alice", 1000L)
        val user = api.login("S001", "pw")
        assertNotNull(user)
        assertEquals("Alice", user!!.nickname)
    }

    @Test
    fun login_wrongPassword_returnsNull() = runTest {
        api.register("S001", "pw", "Alice", 1000L)
        assertNull(api.login("S001", "wrong"))
    }

    @Test
    fun updateProfile_updatesFields() = runTest {
        val id = api.register("S001", "pw", "Alice", 1000L)
        api.updateProfile(id, "New", "138", 2000L)
        val u = api.getUser(id)!!
        assertEquals("New", u.nickname)
        assertEquals("138", u.phone)
        assertEquals(2000L, u.updatedAt)
    }

    @Test
    fun updateAvatar_updatesPath() = runTest {
        val id = api.register("S001", "pw", "Alice", 1000L)
        api.updateAvatar(id, "/img/a.jpg", 2000L)
        assertEquals("/img/a.jpg", api.getUser(id)!!.avatarPath)
    }

    // ── Product ──

    @Test
    fun publish_returnsIdAndStoresData() = runTest {
        val pid = api.publish(1L, "iPhone", "desc", 99900L, "数码", null, "ON_SALE", 1001L)
        assertEquals(1L, pid)
        val p = api.getProduct(pid)!!
        assertEquals("iPhone", p.title)
        assertEquals(99900L, p.priceCents)
    }

    @Test
    fun updateProductStatus_changesStatus() = runTest {
        api.publish(1L, "Phone", "desc", 100L, "数码", null, "ON_SALE", 1000L)
        api.updateProductStatus(1L, "SOLD", 2000L)
        assertEquals("SOLD", api.getProduct(1L)!!.status)
    }

    // ── Favorite ──

    @Test
    fun addAndRemoveFavorite() = runTest {
        api.addFavorite(1L, 10L, 1000L)
        assertEquals(1, api.getFavorites(1L).size)
        api.removeFavorite(1L, 10L)
        assertEquals(0, api.getFavorites(1L).size)
    }

    @Test
    fun addFavorite_idempotent() = runTest {
        api.addFavorite(1L, 10L, 1000L)
        api.addFavorite(1L, 10L, 1001L)
        assertEquals(1, api.getFavorites(1L).size)
    }

    // ── Order ──

    @Test
    fun createOrder_returnsId() = runTest {
        val oid = api.createOrder(10L, 1L, 2L, 0, 1000L, 1000L)
        assertEquals(1L, oid)
        assertEquals(0, api.getOrder(oid)!!.status)
    }

    @Test
    fun updateOrderStatus() = runTest {
        api.createOrder(10L, 1L, 2L, 0, 1000L, 1000L)
        api.updateOrderStatus(1L, 2, 2000L)
        assertEquals(2, api.getOrder(1L)!!.status)
    }

    // ── Chat ──

    @Test
    fun findOrCreateConversation_createsNew() = runTest {
        val cid = api.findOrCreateConversation(10L, 1L, 2L, 1000L)
        assertEquals(1L, cid)
        assertNotNull(api.getConversation(cid))
    }

    @Test
    fun findOrCreateConversation_returnsExisting() = runTest {
        api.findOrCreateConversation(10L, 1L, 2L, 1000L)
        val cid2 = api.findOrCreateConversation(10L, 1L, 2L, 2000L)
        assertEquals(1L, cid2)
    }

    @Test
    fun sendMessage_returnsIdAndStores() = runTest {
        api.findOrCreateConversation(10L, 1L, 2L, 1000L)
        val mid = api.sendMessage(1L, 1L, "你好", 1001L)
        assertEquals(1L, mid)
        val msgs = api.getMessages(1L)
        assertEquals(1, msgs.size)
        assertEquals("你好", msgs[0].content)
    }

    @Test
    fun markConversationRead_marksOtherUsersMessages() = runTest {
        api.findOrCreateConversation(10L, 1L, 2L, 1000L)
        api.sendMessage(1L, 2L, "hi", 1001L)  // 对方发的
        assertEquals(0, api.getMessages(1L)[0].isRead)

        api.markConversationRead(1L, 1L)  // 我标记已读
        assertEquals(1, api.getMessages(1L)[0].isRead)
    }

    @Test
    fun updateConversationLastMessage() = runTest {
        api.findOrCreateConversation(10L, 1L, 2L, 1000L)
        api.updateConversationLastMessage(1L, 5000L)
        assertEquals(5000L, api.getConversation(1L)!!.lastMessageAt)
    }

    // ── Sync ──

    @Test
    fun sync_withoutRoomDb_isNoOp() = runTest {
        // 不传 roomDb，sync 不应抛异常
        api.register("S1", "pw", "A", 0L)
        api.fullSync(1L)
        api.syncUsers()
        api.syncProducts()
        api.syncFavorites(1L)
        api.syncOrders(1L)
        api.syncConversations(1L)
        api.syncMessages(1L)
        api.syncNotifications(1L)
        api.syncProductImages()
        assertTrue(api.called("syncUsers"))
    }

    // ── 辅助方法 ──

    @Test
    fun callTracking() = runTest {
        api.register("S1", "pw", "A", 0L)
        api.login("S1", "pw")
        api.publish(1L, "item", "d", 100L, "cat", null, "ON_SALE", 0L)

        assertTrue(api.called("register"))
        assertTrue(api.called("login"))
        assertTrue(api.called("publish"))
        assertEquals(3, api.calls.size)
    }

    @Test
    fun callCount() = runTest {
        api.register("S1", "pw", "A", 0L)
        api.register("S2", "pw", "B", 0L)
        assertEquals(2, api.callCount("register"))
    }

    @Test
    fun clear_resetsAll() = runTest {
        api.register("S1", "pw", "A", 0L)
        api.clear()
        assertNull(api.getUser(1L))
        assertEquals(0, api.calls.size)
    }

    @Test
    fun putUser_injectsAndAdjustsNextId() = runTest {
        api.putUser(com.example.myapplication.data.db.UserEntity(
            id = 10L, studentId = "S", password = "p", nickname = "n", createdAt = 0L
        ))
        val nextId = api.register("S2", "pw", "B", 0L)
        assertEquals(11L, nextId)  // 自增从 10+1 开始
    }
}