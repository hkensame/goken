package com.example.myapplication.ui

import androidx.compose.runtime.staticCompositionLocalOf
import com.example.myapplication.data.AppContainer

val LocalAppContainer = staticCompositionLocalOf<AppContainer> {
    error("AppContainer not provided")
}
