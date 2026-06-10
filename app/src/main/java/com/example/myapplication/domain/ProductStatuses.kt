package com.example.myapplication.domain

/** 商品上架状态（SQLite 中存字符串） */
object ProductStatus {
    const val ON_SALE = "ON_SALE"
    const val SOLD = "SOLD"
}

/**
 * 订单状态流转（课上演示可画简单状态图）：
 * 待确认 → 交易中 → 已完成
 *        ↘ 已取消
 * 交易中也可取消；完成后不可再改。
 */
object OrderStatus {
    const val PENDING = 0       // 待卖家确认
    const val CONFIRMED = 1   // 已确认 / 交易中
    const val COMPLETED = 2   // 已完成
    const val CANCELLED = 3   // 已取消

    fun label(code: Int): String = when (code) {
        PENDING -> "待确认"
        CONFIRMED -> "交易中"
        COMPLETED -> "已完成"
        CANCELLED -> "已取消"
        else -> "未知"
    }
}

/** 通知类型（SQLite 中存字符串） */
object NotificationType {
    const val ORDER_CONFIRMED = "ORDER_CONFIRMED"   // 卖家确认订单
    const val ORDER_COMPLETED = "ORDER_COMPLETED"   // 交易完成
    const val ORDER_CANCELLED = "ORDER_CANCELLED"   // 订单取消
    const val NEW_MESSAGE = "NEW_MESSAGE"           // 新消息

    fun label(type: String): String = when (type) {
        ORDER_CONFIRMED -> "订单已确认"
        ORDER_COMPLETED -> "交易完成"
        ORDER_CANCELLED -> "订单已取消"
        NEW_MESSAGE -> "新消息"
        else -> type
    }
}
