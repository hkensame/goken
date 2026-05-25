package com.example.myapplication.domain

import org.junit.Assert.assertEquals
import org.junit.Test

/** 纯 JVM 单测：不依赖 Android，也不需要 Mock。 */
class OrderStatusLabelTest {
    @Test
    fun label_mapsKnownCodes() {
        assertEquals("待确认", OrderStatus.label(OrderStatus.PENDING))
        assertEquals("交易中", OrderStatus.label(OrderStatus.CONFIRMED))
        assertEquals("已完成", OrderStatus.label(OrderStatus.COMPLETED))
        assertEquals("已取消", OrderStatus.label(OrderStatus.CANCELLED))
    }

    @Test
    fun label_unknownReturnsFallback() {
        assertEquals("未知", OrderStatus.label(99))
    }
}
