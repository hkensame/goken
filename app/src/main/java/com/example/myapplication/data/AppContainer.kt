package com.example.myapplication.data

import android.content.Context
import com.example.myapplication.data.db.CampusDatabase
import com.example.myapplication.data.remote.RemoteApi
import com.example.myapplication.data.repo.AuthRepository
import com.example.myapplication.data.repo.ChatRepository
import com.example.myapplication.data.repo.FavoriteRepository
import com.example.myapplication.data.repo.OrderRepository
import com.example.myapplication.data.repo.ProductRepository
import com.example.myapplication.data.session.SessionRepository

class AppContainer(
    context: Context,
    val db: CampusDatabase,
) {
    val appContext = context.applicationContext

    val session = SessionRepository(appContext)

    /** 远程 API — 后期注入 Retrofit 实现即可切换为远程优先模式 */
    var api: RemoteApi? = null

    val auth: AuthRepository
        get() = AuthRepository(db.userDao(), session, api)
    val products: ProductRepository
        get() = ProductRepository(db.productDao(), appContext, api)
    val favorites: FavoriteRepository
        get() = FavoriteRepository(db.favoriteDao(), api)
    val orders: OrderRepository
        get() = OrderRepository(db.orderDao(), db.productDao(), api)
    val chat: ChatRepository
        get() = ChatRepository(db, api)
}
