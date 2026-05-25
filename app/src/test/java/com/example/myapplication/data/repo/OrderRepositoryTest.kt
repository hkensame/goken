package com.example.myapplication.data.repo

import com.example.myapplication.data.db.OrderDao
import com.example.myapplication.data.db.OrderEntity
import com.example.myapplication.data.db.ProductDao
import com.example.myapplication.data.db.ProductEntity
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
    private val repo = OrderRepository(orderDao, productDao)

    private fun onSaleProduct(
        id: Long = 100L,
        sellerId: Long = 10L,
    ) = ProductEntity(
        id = id,
        sellerId = sellerId,
        title = "书",
        description = "二手",
        priceCents = 500L,
        category = "教材图书",
        imageLocalPath = null,
        status = ProductStatus.ON_SALE,
        createdAt = 0L,
    )

    @Test
    fun createOrder_productMissing_fails(): Unit = runTest {
        coEvery { productDao.getById(1L) } returns null
        val r = repo.createOrder(productId = 1L, buyerId = 2L)
        assertTrue(r.isFailure)
    }

    @Test
    fun createOrder_notOnSale_fails(): Unit = runTest {
        coEvery { productDao.getById(100L) } returns onSaleProduct().copy(status = ProductStatus.SOLD)
        val r = repo.createOrder(productId = 100L, buyerId = 2L)
        assertTrue(r.isFailure)
    }

    @Test
    fun createOrder_buyOwnProduct_fails(): Unit = runTest {
        coEvery { productDao.getById(100L) } returns onSaleProduct(sellerId = 5L)
        val r = repo.createOrder(productId = 100L, buyerId = 5L)
        assertTrue(r.isFailure)
    }

    @Test
    fun createOrder_openOrderExists_fails(): Unit = runTest {
        coEvery { productDao.getById(100L) } returns onSaleProduct()
        coEvery { orderDao.findOpenOrderForProduct(100L) } returns OrderEntity(
            id = 9L,
            productId = 100L,
            buyerId = 2L,
            sellerId = 10L,
            status = OrderStatus.PENDING,
            createdAt = 0L,
            updatedAt = 0L,
        )
        val r = repo.createOrder(productId = 100L, buyerId = 2L)
        assertTrue(r.isFailure)
    }

    @Test
    fun createOrder_success_insertsPending(): Unit = runTest {
        coEvery { productDao.getById(100L) } returns onSaleProduct()
        coEvery { orderDao.findOpenOrderForProduct(100L) } returns null
        coEvery { orderDao.insert(any()) } returns 77L

        val r = repo.createOrder(productId = 100L, buyerId = 2L)
        assertTrue(r.isSuccess)
        assertEquals(77L, r.getOrNull())

        val captured = slot<OrderEntity>()
        coVerify(exactly = 1) { orderDao.insert(capture(captured)) }
        assertEquals(OrderStatus.PENDING, captured.captured.status)
        assertEquals(2L, captured.captured.buyerId)
        assertEquals(10L, captured.captured.sellerId)
        assertEquals(100L, captured.captured.productId)
    }

    @Test
    fun transition_orderMissing_fails(): Unit = runTest {
        coEvery { orderDao.getById(1L) } returns null
        val r = repo.transition(1L, actorUserId = 2L, newStatus = OrderStatus.CANCELLED)
        assertTrue(r.isFailure)
    }

    @Test
    fun transition_actorNotParty_fails(): Unit = runTest {
        coEvery { orderDao.getById(1L) } returns OrderEntity(
            id = 1L,
            productId = 100L,
            buyerId = 2L,
            sellerId = 10L,
            status = OrderStatus.PENDING,
            createdAt = 0L,
            updatedAt = 0L,
        )
        val r = repo.transition(1L, actorUserId = 999L, newStatus = OrderStatus.CANCELLED)
        assertTrue(r.isFailure)
    }

    @Test
    fun transition_terminalState_fails(): Unit = runTest {
        coEvery { orderDao.getById(1L) } returns OrderEntity(
            id = 1L,
            productId = 100L,
            buyerId = 2L,
            sellerId = 10L,
            status = OrderStatus.COMPLETED,
            createdAt = 0L,
            updatedAt = 0L,
        )
        val r = repo.transition(1L, actorUserId = 2L, newStatus = OrderStatus.CANCELLED)
        assertTrue(r.isFailure)
    }

    @Test
    fun transition_pending_buyerCannotConfirm_fails(): Unit = runTest {
        coEvery { orderDao.getById(1L) } returns OrderEntity(
            id = 1L,
            productId = 100L,
            buyerId = 2L,
            sellerId = 10L,
            status = OrderStatus.PENDING,
            createdAt = 0L,
            updatedAt = 0L,
        )
        val r = repo.transition(1L, actorUserId = 2L, newStatus = OrderStatus.CONFIRMED)
        assertTrue(r.isFailure)
        coVerify(exactly = 0) { orderDao.update(any()) }
    }

    @Test
    fun transition_pending_sellerConfirms(): Unit = runTest {
        val order = OrderEntity(
            id = 1L,
            productId = 100L,
            buyerId = 2L,
            sellerId = 10L,
            status = OrderStatus.PENDING,
            createdAt = 0L,
            updatedAt = 0L,
        )
        coEvery { orderDao.getById(1L) } returns order
        coEvery { orderDao.update(any()) } returns Unit

        val r = repo.transition(1L, actorUserId = 10L, newStatus = OrderStatus.CONFIRMED)
        assertTrue(r.isSuccess)

        val updated = slot<OrderEntity>()
        coVerify(exactly = 1) { orderDao.update(capture(updated)) }
        assertEquals(OrderStatus.CONFIRMED, updated.captured.status)
    }

    @Test
    fun transition_pending_illegalJumpToCompleted_fails(): Unit = runTest {
        coEvery { orderDao.getById(1L) } returns OrderEntity(
            id = 1L,
            productId = 100L,
            buyerId = 2L,
            sellerId = 10L,
            status = OrderStatus.PENDING,
            createdAt = 0L,
            updatedAt = 0L,
        )
        val r = repo.transition(1L, actorUserId = 10L, newStatus = OrderStatus.COMPLETED)
        assertTrue(r.isFailure)
    }

    @Test
    fun transition_confirmed_completed_marksProductSold(): Unit = runTest {
        val order = OrderEntity(
            id = 1L,
            productId = 50L,
            buyerId = 2L,
            sellerId = 10L,
            status = OrderStatus.CONFIRMED,
            createdAt = 0L,
            updatedAt = 0L,
        )
        val product = onSaleProduct(id = 50L)
        coEvery { orderDao.getById(1L) } returns order
        coEvery { orderDao.update(any()) } returns Unit
        coEvery { productDao.getById(50L) } returns product
        coEvery { productDao.update(any()) } returns Unit

        val r = repo.transition(1L, actorUserId = 2L, newStatus = OrderStatus.COMPLETED)
        assertTrue(r.isSuccess)

        val updatedProduct = slot<ProductEntity>()
        coVerify(exactly = 1) { productDao.update(capture(updatedProduct)) }
        assertEquals(ProductStatus.SOLD, updatedProduct.captured.status)
    }

    @Test
    fun transition_completed_whenProductGone_skipsProductUpdate(): Unit = runTest {
        val order = OrderEntity(
            id = 1L,
            productId = 50L,
            buyerId = 2L,
            sellerId = 10L,
            status = OrderStatus.CONFIRMED,
            createdAt = 0L,
            updatedAt = 0L,
        )
        coEvery { orderDao.getById(1L) } returns order
        coEvery { orderDao.update(any()) } returns Unit
        coEvery { productDao.getById(50L) } returns null

        val r = repo.transition(1L, actorUserId = 10L, newStatus = OrderStatus.COMPLETED)
        assertTrue(r.isSuccess)
        coVerify(exactly = 0) { productDao.update(any()) }
    }
}
