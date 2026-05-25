package com.example.myapplication

import android.app.Application
import com.example.myapplication.data.AppContainer
import com.example.myapplication.data.db.CampusDatabase

class CampusApplication : Application() {
    lateinit var container: AppContainer
        private set

    override fun onCreate() {
        super.onCreate()
        val db = CampusDatabase.build(this)
        container = AppContainer(applicationContext, db)
    }
}
