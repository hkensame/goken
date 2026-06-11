package com.example.myapplication

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.runtime.CompositionLocalProvider
import com.example.myapplication.ui.CampusTradeApp
import com.example.myapplication.ui.LocalAppContainer
import com.example.myapplication.ui.theme.MyApplicationTheme

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        android.util.Log.d("TEST", "test debug")

        enableEdgeToEdge()
        val app = application as CampusApplication
        setContent {
            MyApplicationTheme {
                CompositionLocalProvider(LocalAppContainer provides app.container) {
                    CampusTradeApp()
                }
            }
        }
    }
}