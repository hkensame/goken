package com.example.myapplication.data.repo

import com.example.myapplication.data.db.OrderDao
import com.example.myapplication.data.db.OrderEntity
import com.example.myapplication.data.db.ProductDao
import com.example.myapplication.domain.OrderStatus
import com.example.myapplication.domain.ProductStatus
import kotlinx.coroutines.flow.Flow

class OrderRepository(
    private val orderDao: OrderDao,
    private val productDao: ProductDao,
) {
    fun observeMine(userId: Long): Flow<List<com.example.myapplication.data.db.OrderWithDetails>> =
        orderDao.observeMine(userId)

    suspend fun createOrder(productId: Long, buyerId: Long): Result<Long> {
        val product = productDao.getById(productId)
            ?: return Result.failure(IllegalArgumentException("商品不存在"))
        if (product.status != ProductStatus.ON_SALE) {
            return Result.failure(IllegalStateException("商品非在售状态"))
        }
        if (product.sellerId == buyerId) {
            return Result.failure(IllegalStateException("不能购买自己的商品"))
        }
        val open = orderDao.findOpenOrderForProduct(productId)
        if (open != null) {
            return Result.failure(IllegalStateException("该商品已有进行中的订单"))
        }
        val now = System.currentTimeMillis()
        val id = orderDao.insert(
            OrderEntity(
                productId = productId,
                buyerId = buyerId,
                sellerId = product.sellerId,
                status = OrderStatus.PENDING,
                createdAt = now,
                updatedAt = now,
            ),
        )
        return Result.success(id)
    }

    /**
     * 合法流转：
     * - 待确认：买卖双方都可取消；卖家可确认进入交易中
     * - 交易中：买卖双方都可取消；任一方可确认完成
     * - 终态不可再改
     */
    suspend fun transition(orderId: Long, actorUserId: Long, newStatus: Int): Result<Unit> {
        val order = orderDao.getById(orderId)
            ?: return Result.failure(IllegalArgumentException("订单不存在"))
        if (actorUserId != order.buyerId && actorUserId != order.sellerId) {
            return Result.failure(IllegalStateException("无权操作此订单"))
        }
        val old = order.status
        if (old == OrderStatus.COMPLETED || old == OrderStatus.CANCELLED) {
            return Result.failure(IllegalStateException("订单已结束"))
        }
        val ok = when (old) {
            OrderStatus.PENDING -> when (newStatus) {
                OrderStatus.CONFIRMED -> actorUserId == order.sellerId
                OrderStatus.CANCELLED -> true
                else -> false
            }
            OrderStatus.CONFIRMED -> when (newStatus) {
                OrderStatus.COMPLETED -> true
                OrderStatus.CANCELLED -> true
                else -> false
            }
            else -> false
        }
        if (!ok) return Result.failure(IllegalStateException("非法状态流转"))

        val now = System.currentTimeMillis()
        orderDao.update(order.copy(status = newStatus, updatedAt = now))

        if (newStatus == OrderStatus.COMPLETED) {
            val p = productDao.getById(order.productId)
            if (p != null) productDao.update(p.copy(status = ProductStatus.SOLD))
        }
        return Result.success(Unit)
    }
}
