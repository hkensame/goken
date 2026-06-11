package com.example.myapplication

import android.app.Application
import com.example.myapplication.data.AppContainer
import com.example.myapplication.data.db.CampusDatabase
import com.example.myapplication.data.remote.MockRemoteApi

class CampusApplication : Application() {
    lateinit var container: AppContainer
        private set

    override fun onCreate() {
        super.onCreate()
        val db = CampusDatabase.build(this)
        container = AppContainer(applicationContext, db)

        // 开发阶段：使用 MockRemoteApi 模拟后端
        // 后期替换为 container.api = RetrofitApi(retrofit)
        // container.api = MockRemoteApi()
    }
}
