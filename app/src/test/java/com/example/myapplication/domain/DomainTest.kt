package com.example.myapplication.domain

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class DomainTest {

    // ── OrderStatus.label ──

    @Test
    fun orderStatus_label_knownCodes() {
        assertEquals("待确认", OrderStatus.label(OrderStatus.PENDING))
        assertEquals("交易中", OrderStatus.label(OrderStatus.CONFIRMED))
        assertEquals("已完成", OrderStatus.label(OrderStatus.COMPLETED))
        assertEquals("已取消", OrderStatus.label(OrderStatus.CANCELLED))
    }

    @Test
    fun orderStatus_label_unknownReturnsFallback() {
        assertEquals("未知", OrderStatus.label(99))
        assertEquals("未知", OrderStatus.label(-1))
    }

    // ── OrderStatus 常量值 ──

    @Test
    fun orderStatus_constantsAreDistinct() {
        val set = setOf(OrderStatus.PENDING, OrderStatus.CONFIRMED, OrderStatus.COMPLETED, OrderStatus.CANCELLED)
        assertEquals(4, set.size)
    }

    // ── ProductStatus ──

    @Test
    fun productStatus_constantsAreDistinct() {
        val set = setOf(ProductStatus.ON_SALE, ProductStatus.SOLD)
        assertEquals(2, set.size)
    }

    // ── ProductCategories ──

    @Test
    fun productCategories_notEmpty() {
        assertTrue(ProductCategories.all.isNotEmpty())
    }

    @Test
    fun productCategories_containsBookCategory() {
        assertTrue(ProductCategories.all.any { it.contains("教材") })
    }

    @Test
    fun productCategories_noDuplicates() {
        assertEquals(ProductCategories.all.size, ProductCategories.all.toSet().size)
    }

    // ── NotificationType.label ──

    @Test
    fun notificationType_label_knownTypes() {
        assertEquals("订单已确认", NotificationType.label(NotificationType.ORDER_CONFIRMED))
        assertEquals("交易完成", NotificationType.label(NotificationType.ORDER_COMPLETED))
        assertEquals("订单已取消", NotificationType.label(NotificationType.ORDER_CANCELLED))
        assertEquals("新消息", NotificationType.label(NotificationType.NEW_MESSAGE))
    }

    @Test
    fun notificationType_label_unknownReturnsItself() {
        assertEquals("SOMETHING_ELSE", NotificationType.label("SOMETHING_ELSE"))
    }
}