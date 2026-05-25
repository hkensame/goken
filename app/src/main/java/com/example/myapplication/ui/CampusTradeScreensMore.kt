package com.example.myapplication.ui

import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.PickVisualMediaRequest
import androidx.activity.result.contract.ActivityResultContracts.PickVisualMedia
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.FavoriteBorder
import androidx.compose.material.icons.filled.ReceiptLong
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExposedDropdownMenuBox
import androidx.compose.material3.ExposedDropdownMenuDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavHostController
import coil.compose.AsyncImage
import com.example.myapplication.domain.OrderStatus
import com.example.myapplication.domain.ProductCategories
import com.example.myapplication.domain.ProductStatus
import kotlinx.coroutines.launch
import java.io.File

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PublishScreen(nav: NavHostController, userId: Long) {
    val c = LocalAppContainer.current
    val scope = rememberCoroutineScope()
    var title by remember { mutableStateOf("") }
    var desc by remember { mutableStateOf("") }
    var price by remember { mutableStateOf("") }
    var cat by remember { mutableStateOf(ProductCategories.all.first()) }
    var expanded by remember { mutableStateOf(false) }
    var imageUri by remember { mutableStateOf<Uri?>(null) }
    var err by remember { mutableStateOf<String?>(null) }
    var ok by remember { mutableStateOf<String?>(null) }
    val pick = rememberLauncherForActivityResult(PickVisualMedia()) { imageUri = it }
    Scaffold(
        topBar = { CenterAlignedTopAppBar(title = { Text("发布闲置") }) },
    ) { padding ->
        Column(
            Modifier
                .padding(padding)
                .padding(16.dp)
                .fillMaxSize(),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            OutlinedTextField(title, { title = it }, Modifier.fillMaxWidth(), label = { Text("标题") })
            OutlinedTextField(desc, { desc = it }, Modifier.fillMaxWidth(), label = { Text("描述") }, minLines = 3)
            OutlinedTextField(price, { price = it }, Modifier.fillMaxWidth(), label = { Text("价格（元）") }, singleLine = true)
            ExposedDropdownMenuBox(expanded, { expanded = it }) {
                OutlinedTextField(
                    cat,
                    { },
                    Modifier
                        .fillMaxWidth()
                        .menuAnchor(),
                    readOnly = true,
                    label = { Text("分类") },
                    trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded) },
                )
                ExposedDropdownMenu(expanded, { expanded = false }) {
                    ProductCategories.all.forEach { option ->
                        DropdownMenuItem(
                            text = { Text(option) },
                            onClick = {
                                cat = option
                                expanded = false
                            },
                        )
                    }
                }
            }
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                Button(onClick = {
                    pick.launch(PickVisualMediaRequest(PickVisualMedia.ImageOnly))
                }) { Text("选择图片") }
                imageUri?.let { Text("已选图片", style = MaterialTheme.typography.bodySmall) }
            }
            err?.let { Text(it, color = MaterialTheme.colorScheme.error) }
            ok?.let { Text(it, color = MaterialTheme.colorScheme.primary) }
            Button(
                onClick = {
                    scope.launch {
                        err = null
                        ok = null
                        c.products.publish(userId, title, desc, price, cat, imageUri)
                            .onSuccess {
                                ok = "发布成功"
                                title = ""
                                desc = ""
                                price = ""
                                imageUri = null
                            }
                            .onFailure { err = it.message }
                    }
                },
                Modifier.fillMaxWidth(),
            ) { Text("提交") }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProductDetailScreen(nav: NavHostController, userId: Long, productId: Long) {
    val c = LocalAppContainer.current
    val scope = rememberCoroutineScope()
    val product by c.products.observeProduct(productId).collectAsStateWithLifecycle(initialValue = null)
    val fav by c.favorites.observeFavorite(userId, productId).collectAsStateWithLifecycle(initialValue = false)
    var sellerName by remember { mutableStateOf("") }
    LaunchedEffect(product?.sellerId) {
        val sid = product?.sellerId ?: return@LaunchedEffect
        sellerName = c.auth.user(sid)?.nickname.orEmpty()
    }
    var orderMsg by remember { mutableStateOf<String?>(null) }
    Scaffold(
        topBar = {
            CenterAlignedTopAppBar(
                title = { Text("商品详情") },
                navigationIcon = {
                    IconButton(onClick = { nav.popBackStack() }) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, "back")
                    }
                },
                actions = {
                    IconButton(onClick = {
                        scope.launch { c.favorites.toggle(userId, productId) }
                    }) {
                        Icon(if (fav) Icons.Default.Favorite else Icons.Default.FavoriteBorder, "收藏")
                    }
                },
            )
        },
    ) { padding ->
        val p = product
        if (p == null) {
            Box(Modifier.padding(padding).fillMaxSize(), contentAlignment = Alignment.Center) {
                Text("加载中…")
            }
            return@Scaffold
        }
        Column(
            Modifier
                .padding(padding)
                .padding(16.dp)
                .verticalScroll(rememberScrollState()),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            val path = p.imageLocalPath
            if (path != null && File(path).exists()) {
                AsyncImage(
                    File(path),
                    null,
                    Modifier
                        .fillMaxWidth()
                        .height(220.dp),
                    contentScale = ContentScale.Crop,
                )
            }
            Text(p.title, style = MaterialTheme.typography.headlineSmall)
            Text("¥${"%.2f".format(p.priceCents / 100.0)} · ${p.category} · ${p.status}")
            Text("卖家：$sellerName")
            Text(p.description)
            orderMsg?.let { Text(it, color = MaterialTheme.colorScheme.error) }
            if (p.sellerId != userId && p.status == ProductStatus.ON_SALE) {
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    OutlinedButton(onClick = {
                        scope.launch {
                            val cid = c.chat.getOrCreateConversation(productId, userId, p.sellerId)
                            nav.navigate("chat/$cid")
                        }
                    }) { Text("联系卖家") }
                    Button(onClick = {
                        scope.launch {
                            c.orders.createOrder(productId, userId)
                                .onSuccess { orderMsg = "订单已创建，待卖家确认" }
                                .onFailure { orderMsg = it.message }
                        }
                    }) { Text("下单购买") }
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FavoritesScreen(nav: NavHostController, userId: Long) {
    val c = LocalAppContainer.current
    val items by c.favorites.observeFavorites(userId).collectAsStateWithLifecycle(initialValue = emptyList())
    Scaffold(topBar = { CenterAlignedTopAppBar(title = { Text("我的收藏") }) }) { padding ->
        LazyColumn(
            Modifier.padding(padding).fillMaxSize(),
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            items(items, key = { it.id }) { p ->
                Card(
                    Modifier
                        .fillMaxWidth()
                        .clickable { nav.navigate("product/${p.id}") },
                ) {
                    Text(p.title, Modifier.padding(16.dp), style = MaterialTheme.typography.titleMedium)
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ChatListScreen(nav: NavHostController, userId: Long) {
    val c = LocalAppContainer.current
    val list by c.chat.observeConversations(userId).collectAsStateWithLifecycle(initialValue = emptyList())
    Scaffold(topBar = { CenterAlignedTopAppBar(title = { Text("消息") }) }) { padding ->
        LazyColumn(Modifier.padding(padding).fillMaxSize(), contentPadding = PaddingValues(16.dp)) {
            items(list, key = { it.conversationId }) { row ->
                Card(
                    Modifier
                        .fillMaxWidth()
                        .padding(bottom = 8.dp)
                        .clickable { nav.navigate("chat/${row.conversationId}") },
                ) {
                    Column(Modifier.padding(12.dp)) {
                        Text("关于：${row.productTitle}", style = MaterialTheme.typography.titleSmall)
                        Text("对方：${row.peerNickname}")
                        row.lastMessagePreview?.let { Text(it, style = MaterialTheme.typography.bodySmall) }
                    }
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class, ExperimentalFoundationApi::class)
@Composable
fun ChatThreadScreen(nav: NavHostController, userId: Long, conversationId: Long) {
    val c = LocalAppContainer.current
    val scope = rememberCoroutineScope()
    val messages by c.chat.observeMessages(conversationId).collectAsStateWithLifecycle(initialValue = emptyList())
    var input by remember { mutableStateOf("") }
    val listState = rememberLazyListState()
    LaunchedEffect(messages.size) {
        if (messages.isNotEmpty()) listState.animateScrollToItem(messages.lastIndex)
    }
    Scaffold(
        topBar = {
            CenterAlignedTopAppBar(
                title = { Text("聊天") },
                navigationIcon = {
                    IconButton(onClick = { nav.popBackStack() }) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, "back")
                    }
                },
            )
        },
    ) { padding ->
        Column(
            Modifier
                .padding(padding)
                .fillMaxSize()
                .navigationBarsPadding()
                .imePadding(),
        ) {
            LazyColumn(
                Modifier
                    .weight(1f)
                    .fillMaxWidth()
                    .padding(horizontal = 12.dp),
                state = listState,
                verticalArrangement = Arrangement.spacedBy(8.dp),
                contentPadding = PaddingValues(vertical = 12.dp),
            ) {
                items(messages, key = { it.id }) { m ->
                    val mine = m.senderId == userId
                    Row(
                        Modifier.fillMaxWidth(),
                        horizontalArrangement = if (mine) Arrangement.End else Arrangement.Start,
                    ) {
                        Card {
                            Text(m.content, Modifier.padding(12.dp))
                        }
                    }
                }
            }
            Row(
                Modifier
                    .fillMaxWidth()
                    .padding(8.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                OutlinedTextField(input, { input = it }, Modifier.weight(1f), placeholder = { Text("消息…") })
                Button(onClick = {
                    val t = input
                    input = ""
                    scope.launch { c.chat.sendMessage(conversationId, userId, t) }
                }) { Text("发送") }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun OrdersScreen(nav: NavHostController, userId: Long) {
    val c = LocalAppContainer.current
    val scope = rememberCoroutineScope()
    val orders by c.orders.observeMine(userId).collectAsStateWithLifecycle(initialValue = emptyList())
    Scaffold(
        topBar = {
            CenterAlignedTopAppBar(
                title = { Text("我的订单") },
                navigationIcon = {
                    IconButton(onClick = { nav.popBackStack() }) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, "back")
                    }
                },
            )
        },
    ) { padding ->
        LazyColumn(
            Modifier.padding(padding).fillMaxSize(),
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            items(orders, key = { it.orderId }) { o ->
                Card(Modifier.fillMaxWidth()) {
                    Column(Modifier.padding(12.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                        Text(o.productTitle, style = MaterialTheme.typography.titleSmall)
                        Text("金额 ¥${"%.2f".format(o.priceCents / 100.0)}")
                        Text("状态：${OrderStatus.label(o.status)}")
                        Text(
                            if (o.buyerId == userId) "我是买家 · 卖家 ${o.sellerNickname}"
                            else "我是卖家 · 买家 ${o.buyerNickname}",
                            style = MaterialTheme.typography.bodySmall,
                        )
                        OrderActionRow(o, userId) { newStatus ->
                            scope.launch { c.orders.transition(o.orderId, userId, newStatus) }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun OrderActionRow(
    o: com.example.myapplication.data.db.OrderWithDetails,
    userId: Long,
    onTransition: (Int) -> Unit,
) {
    val st = o.status
    Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
        if (st == OrderStatus.PENDING) {
            if (o.sellerId == userId) {
                TextButton(onClick = { onTransition(OrderStatus.CONFIRMED) }) { Text("确认订单") }
            }
            TextButton(onClick = { onTransition(OrderStatus.CANCELLED) }) { Text("取消") }
        }
        if (st == OrderStatus.CONFIRMED) {
            TextButton(onClick = { onTransition(OrderStatus.COMPLETED) }) { Text("完成交易") }
            TextButton(onClick = { onTransition(OrderStatus.CANCELLED) }) { Text("取消") }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProfileScreen(nav: NavHostController, userId: Long) {
    val c = LocalAppContainer.current
    val scope = rememberCoroutineScope()
    val user by c.db.userDao().observeUser(userId).collectAsStateWithLifecycle(initialValue = null)
    Scaffold(topBar = { CenterAlignedTopAppBar(title = { Text("我的") }) }) { padding ->
        Column(
            Modifier
                .padding(padding)
                .padding(24.dp)
                .fillMaxSize(),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            Text("学号：${user?.studentId.orEmpty()}")
            Text("昵称：${user?.nickname.orEmpty()}")
            Button(onClick = { nav.navigate("orders") }, Modifier.fillMaxWidth()) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(Icons.Default.ReceiptLong, contentDescription = null)
                    Spacer(Modifier.width(8.dp))
                    Text("我的订单")
                }
            }
            OutlinedButton(
                onClick = {
                    scope.launch {
                        c.auth.logout()
                    }
                },
                Modifier.fillMaxWidth(),
            ) { Text("退出登录") }
        }
    }
}
