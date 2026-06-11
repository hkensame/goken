package com.example.myapplication.data.repo

import com.example.myapplication.data.db.OrderDao
import com.example.myapplication.data.db.OrderEntity
import com.example.myapplication.data.db.ProductDao
import com.example.myapplication.data.db.ProductEntity
import com.example.myapplication.data.remote.MockRemoteApi
import com.example.myapplication.domain.OrderStatus
import com.example.myapplication.domain.ProductStatus
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import io.mockk.slot
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class OrderRepositoryTest {

    private val orderDao = mockk<OrderDao>()
    private val productDao = mockk<ProductDao>()
    private val roomRepo = OrderRepository(orderDao, productDao)

    private fun onSaleProduct(id: Long = 100L, sellerId: Long = 10L) = ProductEntity(
        id = id, sellerId = sellerId, title = "书", description = "二手",
        priceCents = 500L, category = "教材图书", imageLocalPath = null,
        status = ProductStatus.ON_SALE, createdAt = 0L,
    )

    // ════════════════════════════════════════
    //  Room 分支 — createOrder
    // ════════════════════════════════════════

    @Test
    fun room_createOrder_productMissing_fails() = runTest {
        coEvery { productDao.getById(1L) } returns null
        assertTrue(roomRepo.createOrder(1L, 2L).isFailure)
    }

    @Test
    fun room_createOrder_notOnSale_fails() = runTest {
        coEvery { productDao.getById(100L) } returns onSaleProduct().copy(status = ProductStatus.SOLD)
        assertTrue(roomRepo.createOrder(100L, 2L).isFailure)
    }

    @Test
    fun room_createOrder_buyOwnProduct_fails() = runTest {
        coEvery { productDao.getById(100L) } returns onSaleProduct(sellerId = 5L)
        assertTrue(roomRepo.createOrder(100L, 5L).isFailure)
    }

    @Test
    fun room_createOrder_openOrderExists_fails() = runTest {
        coEvery { productDao.getById(100L) } returns onSaleProduct()
        coEvery { orderDao.findOpenOrderForProduct(100L) } returns OrderEntity(
            id = 9L, productId = 100L, buyerId = 2L, sellerId = 10L,
            status = OrderStatus.PENDING, createdAt = 0L, updatedAt = 0L,
        )
        assertTrue(roomRepo.createOrder(100L, 2L).isFailure)
    }

    @Test
    fun room_createOrder_success() = runTest {
        coEvery { productDao.getById(100L) } returns onSaleProduct()
        coEvery { orderDao.findOpenOrderForProduct(100L) } returns null
        coEvery { orderDao.insert(any()) } returns 77L

        val r = roomRepo.createOrder(100L, 2L)
        assertTrue(r.isSuccess)
        assertEquals(77L, r.getOrNull())

        val captured = slot<OrderEntity>()
        coVerify { orderDao.insert(capture(captured)) }
        assertEquals(OrderStatus.PENDING, captured.captured.status)
        assertEquals(2L, captured.captured.buyerId)
        assertEquals(10L, captured.captured.sellerId)
    }

    // ════════════════════════════════════════
    //  Room 分支 — transition
    // ════════════════════════════════════════

    @Test
    fun room_transition_orderMissing_fails() = runTest {
        coEvery { orderDao.getById(1L) } returns null
        assertTrue(roomRepo.transition(1L, 2L, OrderStatus.CANCELLED).isFailure)
    }

    @Test
    fun room_transition_actorNotParty_fails() = runTest {
        coEvery { orderDao.getById(1L) } returns OrderEntity(
            id = 1L, productId = 100L, buyerId = 2L, sellerId = 10L,
            status = OrderStatus.PENDING, createdAt = 0L, updatedAt = 0L,
        )
        assertTrue(roomRepo.transition(1L, 999L, OrderStatus.CANCELLED).isFailure)
    }

    @Test
    fun room_transition_terminalState_fails() = runTest {
        coEvery { orderDao.getById(1L) } returns OrderEntity(
            id = 1L, productId = 100L, buyerId = 2L, sellerId = 10L,
            status = OrderStatus.COMPLETED, createdAt = 0L, updatedAt = 0L,
        )
        assertTrue(roomRepo.transition(1L, 2L, OrderStatus.CANCELLED).isFailure)
    }

    @Test
    fun room_transition_pending_buyerCannotConfirm() = runTest {
        coEvery { orderDao.getById(1L) } returns OrderEntity(
            id = 1L, productId = 100L, buyerId = 2L, sellerId = 10L,
            status = OrderStatus.PENDING, createdAt = 0L, updatedAt = 0L,
        )
        assertTrue(roomRepo.transition(1L, 2L, OrderStatus.CONFIRMED).isFailure)
        coVerify(exactly = 0) { orderDao.update(any()) }
    }

    @Test
    fun room_transition_pending_sellerConfirms() = runTest {
        coEvery { orderDao.getById(1L) } returns OrderEntity(
            id = 1L, productId = 100L, buyerId = 2L, sellerId = 10L,
            status = OrderStatus.PENDING, createdAt = 0L, updatedAt = 0L,
        )
        coEvery { orderDao.update(any()) } returns Unit

        val r = roomRepo.transition(1L, 10L, OrderStatus.CONFIRMED)
        assertTrue(r.isSuccess)

        val updated = slot<OrderEntity>()
        coVerify { orderDao.update(capture(updated)) }
        assertEquals(OrderStatus.CONFIRMED, updated.captured.status)
    }

    @Test
    fun room_transition_confirmed_completed_marksProductSold() = runTest {
        coEvery { orderDao.getById(1L) } returns OrderEntity(
            id = 1L, productId = 50L, buyerId = 2L, sellerId = 10L,
            status = OrderStatus.CONFIRMED, createdAt = 0L, updatedAt = 0L,
        )
        coEvery { orderDao.update(any()) } returns Unit
        coEvery { productDao.getById(50L) } returns onSaleProduct(id = 50L)
        coEvery { productDao.update(any()) } returns Unit

        val r = roomRepo.transition(1L, 2L, OrderStatus.COMPLETED)
        assertTrue(r.isSuccess)

        val updatedProduct = slot<ProductEntity>()
        coVerify { productDao.update(capture(updatedProduct)) }
        assertEquals(ProductStatus.SOLD, updatedProduct.captured.status)
    }

    @Test
    fun room_transition_confirmed_cancelled_byEither() = runTest {
        coEvery { orderDao.getById(1L) } returns OrderEntity(
            id = 1L, productId = 100L, buyerId = 2L, sellerId = 10L,
            status = OrderStatus.CONFIRMED, createdAt = 0L, updatedAt = 0L,
        )
        coEvery { orderDao.update(any()) } returns Unit

        // 买家取消
        assertTrue(roomRepo.transition(1L, 2L, OrderStatus.CANCELLED).isSuccess)
    }

    // ════════════════════════════════════════
    //  MockRemoteApi 分支
    // ════════════════════════════════════════

    @Test
    fun api_createOrder_success() = runTest {
        val api = MockRemoteApi()
        coEvery { productDao.getById(100L) } returns onSaleProduct()
        coEvery { orderDao.findOpenOrderForProduct(100L) } returns null
        val repo = OrderRepository(orderDao, productDao, api)

        val r = repo.createOrder(100L, 2L)
        assertTrue(r.isSuccess)
        assertTrue(api.called("createOrder"))
        assertTrue(api.called("syncOrders"))
        // mock 成功不走 Room
        coVerify(exactly = 0) { orderDao.insert(any()) }
    }

    @Test
    fun api_createOrder_productMissing_stillFails() = runTest {
        val api = MockRemoteApi()
        coEvery { productDao.getById(1L) } returns null
        val repo = OrderRepository(orderDao, productDao, api)

        assertTrue(repo.createOrder(1L, 2L).isFailure)
        // 校验失败在 api 之前，不调 mock
        assertEquals(0, api.calls.size)
    }

    @Test
    fun api_transition_completed_callsUpdateOrderAndProduct() = runTest {
        val api = MockRemoteApi()
        coEvery { orderDao.getById(1L) } returns OrderEntity(
            id = 1L, productId = 50L, buyerId = 2L, sellerId = 10L,
            status = OrderStatus.CONFIRMED, createdAt = 0L, updatedAt = 0L,
        )
        val repo = OrderRepository(orderDao, productDao, api)

        val r = repo.transition(1L, 2L, OrderStatus.COMPLETED)
        assertTrue(r.isSuccess)
        assertTrue(api.called("updateOrderStatus"))
        assertTrue(api.called("updateProductStatus"))
        assertTrue(api.called("syncProducts"))
        // mock 成功不走 Room
        coVerify(exactly = 0) { orderDao.update(any()) }
    }
}