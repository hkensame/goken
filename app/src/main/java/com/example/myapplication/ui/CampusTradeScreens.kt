package com.example.myapplication.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Message
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilterChip
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavHostController
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import coil.compose.AsyncImage
import com.example.myapplication.domain.ProductCategories
import kotlinx.coroutines.launch
import java.io.File

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AuthFlow() {
    val nav = rememberNavController()
    NavHost(nav, startDestination = "login") {
        composable("login") { LoginScreen(nav) }
        composable("register") { RegisterScreen(nav) }
    }
}

@Composable
fun LoginScreen(nav: NavHostController) {
    val c = LocalAppContainer.current
    val scope = rememberCoroutineScope()
    var sid by remember { mutableStateOf("") }
    var pwd by remember { mutableStateOf("") }
    var err by remember { mutableStateOf<String?>(null) }
    Column(
        Modifier
            .fillMaxSize()
            .padding(24.dp),
        verticalArrangement = Arrangement.Center,
    ) {
        Text("登录", style = MaterialTheme.typography.headlineMedium)
        Spacer(Modifier.height(16.dp))
        OutlinedTextField(sid, { sid = it }, Modifier.fillMaxWidth(), label = { Text("学号") })
        Spacer(Modifier.height(8.dp))
        OutlinedTextField(pwd, { pwd = it }, Modifier.fillMaxWidth(), label = { Text("密码") })
        err?.let { Text(it, color = MaterialTheme.colorScheme.error) }
        Spacer(Modifier.height(16.dp))
        Button(
            onClick = {
                scope.launch {
                    c.auth.login(sid, pwd).onFailure { err = it.message }
                }
            },
            Modifier.fillMaxWidth(),
        ) { Text("登录") }
        TextButton(onClick = { nav.navigate("register") }, Modifier.fillMaxWidth()) {
            Text("没有账号？去注册")
        }
    }
}

@Composable
fun RegisterScreen(nav: NavHostController) {
    val c = LocalAppContainer.current
    val scope = rememberCoroutineScope()
    var sid by remember { mutableStateOf("") }
    var pwd by remember { mutableStateOf("") }
    var nick by remember { mutableStateOf("") }
    var err by remember { mutableStateOf<String?>(null) }
    Column(Modifier.fillMaxSize().padding(24.dp)) {
        Text("注册", style = MaterialTheme.typography.headlineMedium)
        Spacer(Modifier.height(16.dp))
        OutlinedTextField(sid, { sid = it }, Modifier.fillMaxWidth(), label = { Text("学号") })
        Spacer(Modifier.height(8.dp))
        OutlinedTextField(pwd, { pwd = it }, Modifier.fillMaxWidth(), label = { Text("密码") })
        Spacer(Modifier.height(8.dp))
        OutlinedTextField(nick, { nick = it }, Modifier.fillMaxWidth(), label = { Text("昵称（可选）") })
        err?.let { Text(it, color = MaterialTheme.colorScheme.error) }
        Spacer(Modifier.height(16.dp))
        Button(
            onClick = {
                scope.launch {
                    c.auth.register(sid, pwd, nick).onFailure { err = it.message }
                }
            },
            Modifier.fillMaxWidth(),
        ) { Text("注册并登录") }
        TextButton(onClick = { nav.popBackStack() }) { Text("返回登录") }
    }
}

private val TopLevel = setOf("home", "publish", "favorites", "messages", "profile")

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MainShell(userId: Long) {
    val nav = rememberNavController()
    val backStack by nav.currentBackStackEntryAsState()
    val route = backStack?.destination?.route ?: ""
    val showBar = route in TopLevel
    Scaffold(
        bottomBar = {
            if (!showBar) return@Scaffold
            NavigationBar {
                NavigationBarItem(
                    selected = route == "home",
                    onClick = {
                        nav.navigate("home") {
                            launchSingleTop = true
                            restoreState = true
                        }
                    },
                    icon = { Icon(Icons.Default.Home, null) },
                    label = { Text("首页") },
                )
                NavigationBarItem(
                    selected = route == "publish",
                    onClick = {
                        nav.navigate("publish") {
                            launchSingleTop = true
                            restoreState = true
                        }
                    },
                    icon = { Icon(Icons.Default.Add, null) },
                    label = { Text("发布") },
                )
                NavigationBarItem(
                    selected = route == "favorites",
                    onClick = {
                        nav.navigate("favorites") {
                            launchSingleTop = true
                            restoreState = true
                        }
                    },
                    icon = { Icon(Icons.Default.Favorite, null) },
                    label = { Text("收藏") },
                )
                NavigationBarItem(
                    selected = route == "messages",
                    onClick = {
                        nav.navigate("messages") {
                            launchSingleTop = true
                            restoreState = true
                        }
                    },
                    icon = { Icon(Icons.AutoMirrored.Filled.Message, null) },
                    label = { Text("消息") },
                )
                NavigationBarItem(
                    selected = route == "profile",
                    onClick = {
                        nav.navigate("profile") {
                            launchSingleTop = true
                            restoreState = true
                        }
                    },
                    icon = { Icon(Icons.Default.Person, null) },
                    label = { Text("我的") },
                )
            }
        },
    ) { padding ->
        NavHost(nav, startDestination = "home", Modifier.padding(padding)) {
            composable("home") { HomeScreen(nav, userId) }
            composable("publish") { PublishScreen(nav, userId) }
            composable("favorites") { FavoritesScreen(nav, userId) }
            composable("messages") { ChatListScreen(nav, userId) }
            composable("profile") { ProfileScreen(nav, userId) }
            composable("orders") { OrdersScreen(nav, userId) }
            composable(
                "product/{productId}",
                listOf(navArgument("productId") { type = NavType.LongType }),
            ) {
                ProductDetailScreen(nav, userId, it.arguments!!.getLong("productId"))
            }
            composable(
                "chat/{conversationId}",
                listOf(navArgument("conversationId") { type = NavType.LongType }),
            ) {
                ChatThreadScreen(nav, userId, it.arguments!!.getLong("conversationId"))
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(nav: NavHostController, userId: Long) {
    val c = LocalAppContainer.current
    var search by remember { mutableStateOf("") }
    var category by remember { mutableStateOf<String?>(null) }
    val items by c.products.observeMarket(
        search = search,
        category = category,
        categoryAll = category == null,
    ).collectAsStateWithLifecycle(initialValue = emptyList())
    Scaffold(
        topBar = {
            CenterAlignedTopAppBar(title = { Text("校园二手") })
        },
    ) { padding ->
        Column(Modifier.padding(padding).fillMaxSize()) {
            OutlinedTextField(
                search,
                { search = it },
                Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 8.dp),
                leadingIcon = { Icon(Icons.Default.Search, null) },
                placeholder = { Text("搜索标题或描述") },
                singleLine = true,
            )
            Row(
                Modifier.padding(horizontal = 8.dp),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                FilterChip(
                    category == null,
                    { category = null },
                    { Text("全部分类") },
                )
                ProductCategories.all.forEach { cat ->
                    FilterChip(
                        category == cat,
                        { category = cat },
                        { Text(cat) },
                    )
                }
            }
            LazyColumn(contentPadding = PaddingValues(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                items(items, key = { it.id }) { p ->
                    Card(
                        Modifier
                            .fillMaxWidth()
                            .clickable { nav.navigate("product/${p.id}") },
                    ) {
                        Row(Modifier.padding(12.dp)) {
                            val path = p.imageLocalPath
                            if (path != null && File(path).exists()) {
                                AsyncImage(
                                    File(path),
                                    null,
                                    Modifier
                                        .width(88.dp)
                                        .height(88.dp),
                                    contentScale = ContentScale.Crop,
                                )
                            } else {
                                Spacer(Modifier.width(88.dp).height(88.dp))
                            }
                            Spacer(Modifier.width(12.dp))
                            Column {
                                Text(p.title, style = MaterialTheme.typography.titleMedium)
                                Text("¥${"%.2f".format(p.priceCents / 100.0)} · ${p.category}")
                                Text("卖家：${p.sellerNickname}", style = MaterialTheme.typography.bodySmall)
                            }
                        }
                    }
                }
            }
        }
    }
}
