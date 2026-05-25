package com.example.myapplication.ui

import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.lifecycle.compose.collectAsStateWithLifecycle

@Composable
fun CampusTradeApp() {
    val container = LocalAppContainer.current
    val uid by container.session.userId.collectAsStateWithLifecycle(initialValue = 0L)
    if (uid == 0L) AuthFlow() else MainShell(userId = uid)
}
