package com.example.myapplication.ui

import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.PickVisualMediaRequest
import androidx.activity.result.contract.ActivityResultContracts.PickVisualMedia
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.FavoriteBorder
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.ReceiptLong
import androidx.compose.material.icons.filled.Storefront
import androidx.compose.material3.Badge
import androidx.compose.material3.BadgedBox
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.CircularProgressIndicator
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
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavHostController
import coil.compose.AsyncImage
import com.example.myapplication.data.db.NotificationEntity
import com.example.myapplication.domain.NotificationType
import com.example.myapplication.domain.OrderStatus
import com.example.myapplication.domain.ProductCategories
import com.example.myapplication.domain.ProductStatus
import kotlinx.coroutines.launch
import java.io.File
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

private fun formatTime(millis: Long): String {
    if (millis <= 0) return ""
    val sdf = SimpleDateFormat("yyyy-MM-dd HH:mm", Locale.getDefault())
    return sdf.format(Date(millis))
}

// ─────────────────────────────── 发布 ───────────────────────────────

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
                .fillMaxSize()
                .verticalScroll(rememberScrollState()),
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
            Row(
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Button(onClick = {
                    pick.launch(PickVisualMediaRequest(PickVisualMedia.ImageOnly))
                }) { Text("选择图片") }
                imageUri?.let {
                    Text("已选图片", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.primary)
                }
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

// ─────────────────────────────── 商品详情 ───────────────────────────────

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProductDetailScreen(nav: NavHostController, userId: Long, productId: Long) {
    val c = LocalAppContainer.current
    val scope = rememberCoroutineScope()
    val product by c.products.observeProduct(productId).collectAsStateWithLifecycle(initialValue = null)
    val images by c.db.productImageDao().observeByProduct(productId)
        .collectAsStateWithLifecycle(initialValue = emptyList())
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
                CircularProgressIndicator()
            }
            return@Scaffold
        }
        Column(
            Modifier
                .padding(padding)
                .verticalScroll(rememberScrollState()),
        ) {
            // 多图轮播
            val allImages = buildList {
                p.imageLocalPath?.let { add(it) }
                images.forEach { add(it.imagePath) }
            }.filter { File(it).exists() }

            if (allImages.isNotEmpty()) {
                val pagerState = rememberPagerState(pageCount = { allImages.size })
                Box {
                    HorizontalPager(
                        state = pagerState,
                        Modifier
                            .fillMaxWidth()
                            .height(260.dp),
                    ) { page ->
                        AsyncImage(
                            File(allImages[page]),
                            null,
                            Modifier.fillMaxSize(),
                            contentScale = ContentScale.Crop,
                        )
                    }
                    // 页码指示器
                    if (allImages.size > 1) {
                        Row(
                            Modifier
                                .align(Alignment.BottomCenter)
                                .padding(bottom = 12.dp),
                            horizontalArrangement = Arrangement.spacedBy(6.dp),
                        ) {
                            repeat(allImages.size) { i ->
                                Box(
                                    Modifier
                                        .size(8.dp)
                                        .clip(CircleShape)
                                        .background(
                                            if (i == pagerState.currentPage)
                                                MaterialTheme.colorScheme.primary
                                            else
                                                MaterialTheme.colorScheme.onSurface.copy(alpha = 0.3f),
                                        ),
                                )
                            }
                        }
                    }
                }
            }

            Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Text(p.title, style = MaterialTheme.typography.headlineSmall)
                Text(
                    "¥${"%.2f".format(p.priceCents / 100.0)}",
                    style = MaterialTheme.typography.titleLarge,
                    color = MaterialTheme.colorScheme.primary,
                )
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    Text(p.category, style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    Text("·", color = MaterialTheme.colorScheme.onSurfaceVariant)
                    Text(
                        if (p.status == ProductStatus.ON_SALE) "在售" else "已售",
                        style = MaterialTheme.typography.bodyMedium,
                        color = if (p.status == ProductStatus.ON_SALE)
                            MaterialTheme.colorScheme.primary
                        else
                            MaterialTheme.colorScheme.error,
                    )
                }
                Text("卖家：$sellerName", style = MaterialTheme.typography.bodyMedium)
                if (p.updatedAt > 0) {
                    Text("更新：${formatTime(p.updatedAt)}", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
                Spacer(Modifier.height(4.dp))
                Text(p.description, style = MaterialTheme.typography.bodyLarge)
                orderMsg?.let { Text(it, color = MaterialTheme.colorScheme.error) }
                if (p.sellerId != userId && p.status == ProductStatus.ON_SALE) {
                    Spacer(Modifier.height(4.dp))
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
}

// ─────────────────────────────── 收藏 ───────────────────────────────

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FavoritesScreen(nav: NavHostController, userId: Long) {
    val c = LocalAppContainer.current
    val items by c.favorites.observeFavorites(userId).collectAsStateWithLifecycle(initialValue = emptyList())
    Scaffold(topBar = { CenterAlignedTopAppBar(title = { Text("我的收藏") }) }) { padding ->
        if (items.isEmpty()) {
            Box(Modifier.padding(padding).fillMaxSize(), contentAlignment = Alignment.Center) {
                Text("还没有收藏商品", color = MaterialTheme.colorScheme.onSurfaceVariant)
            }
            return@Scaffold
        }
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
                    colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant),
                ) {
                    Row(Modifier.padding(12.dp)) {
                        val path = p.imageLocalPath
                        if (path != null && File(path).exists()) {
                            AsyncImage(
                                File(path), null,
                                Modifier.size(72.dp).clip(MaterialTheme.shapes.small),
                                contentScale = ContentScale.Crop,
                            )
                        } else {
                            Spacer(Modifier.size(72.dp))
                        }
                        Spacer(Modifier.width(12.dp))
                        Column(Modifier.weight(1f)) {
                            Text(p.title, style = MaterialTheme.typography.titleSmall, maxLines = 1, overflow = TextOverflow.Ellipsis)
                            Text("¥${"%.2f".format(p.priceCents / 100.0)}", color = MaterialTheme.colorScheme.primary)
                            Text(p.sellerNickname, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                        }
                    }
                }
            }
        }
    }
}

// ─────────────────────────────── 会话列表 ───────────────────────────────

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ChatListScreen(nav: NavHostController, userId: Long) {
    val c = LocalAppContainer.current
    val list by c.chat.observeConversations(userId).collectAsStateWithLifecycle(initialValue = emptyList())
    Scaffold(topBar = { CenterAlignedTopAppBar(title = { Text("消息") }) }) { padding ->
        if (list.isEmpty()) {
            Box(Modifier.padding(padding).fillMaxSize(), contentAlignment = Alignment.Center) {
                Text("暂无消息", color = MaterialTheme.colorScheme.onSurfaceVariant)
            }
            return@Scaffold
        }
        LazyColumn(Modifier.padding(padding).fillMaxSize(), contentPadding = PaddingValues(16.dp)) {
            items(list, key = { it.conversationId }) { row ->
                Card(
                    Modifier
                        .fillMaxWidth()
                        .padding(bottom = 8.dp)
                        .clickable { nav.navigate("chat/${row.conversationId}") },
                ) {
                    Row(
                        Modifier.padding(12.dp),
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Column(Modifier.weight(1f)) {
                            Text("关于：${row.productTitle}", style = MaterialTheme.typography.titleSmall)
                            Text("对方：${row.peerNickname}", style = MaterialTheme.typography.bodySmall)
                            row.lastMessagePreview?.let {
                                Text(
                                    it,
                                    style = MaterialTheme.typography.bodySmall,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                                    maxLines = 1,
                                    overflow = TextOverflow.Ellipsis,
                                )
                            }
                        }
                        if (row.unreadCount > 0) {
                            Badge(containerColor = MaterialTheme.colorScheme.error) {
                                Text(
                                    if (row.unreadCount > 99) "99+" else row.unreadCount.toString(),
                                    color = MaterialTheme.colorScheme.onError,
                                )
                            }
                        }
                    }
                }
            }
        }
    }
}

// ─────────────────────────────── 聊天 ───────────────────────────────

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ChatThreadScreen(nav: NavHostController, userId: Long, conversationId: Long) {
    val c = LocalAppContainer.current
    val scope = rememberCoroutineScope()
    val messages by c.chat.observeMessages(conversationId).collectAsStateWithLifecycle(initialValue = emptyList())
    var input by remember { mutableStateOf("") }
    val listState = rememberLazyListState()

    // 进入聊天时标记已读
    LaunchedEffect(conversationId) {
        c.chat.markAsRead(conversationId, userId)
    }
    // 新消息时自动标记已读 + 滚动到底部
    LaunchedEffect(messages.size) {
        if (messages.isNotEmpty()) listState.animateScrollToItem(messages.lastIndex)
        c.chat.markAsRead(conversationId, userId)
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
                        Card(
                            colors = CardDefaults.cardColors(
                                containerColor = if (mine)
                                    MaterialTheme.colorScheme.primaryContainer
                                else
                                    MaterialTheme.colorScheme.surfaceVariant,
                            ),
                        ) {
                            Text(
                                m.content,
                                Modifier.padding(12.dp),
                                color = if (mine)
                                    MaterialTheme.colorScheme.onPrimaryContainer
                                else
                                    MaterialTheme.colorScheme.onSurfaceVariant,
                            )
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

// ─────────────────────────────── 订单 ───────────────────────────────

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
        if (orders.isEmpty()) {
            Box(Modifier.padding(padding).fillMaxSize(), contentAlignment = Alignment.Center) {
                Text("暂无订单", color = MaterialTheme.colorScheme.onSurfaceVariant)
            }
            return@Scaffold
        }
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
                        Text("更新：${formatTime(o.updatedAt)}", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
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

// ─────────────────────────────── 个人中心 ───────────────────────────────

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProfileScreen(nav: NavHostController, userId: Long) {
    val c = LocalAppContainer.current
    val scope = rememberCoroutineScope()
    val user by c.db.userDao().observeUser(userId).collectAsStateWithLifecycle(initialValue = null)
    val unreadNotif by c.db.notificationDao().observeUnreadCount(userId)
        .collectAsStateWithLifecycle(initialValue = 0)

    Scaffold(topBar = { CenterAlignedTopAppBar(title = { Text("我的") }) }) { padding ->
        Column(
            Modifier
                .padding(padding)
                .padding(24.dp)
                .fillMaxSize(),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            // 头像 + 基本信息
            Row(verticalAlignment = Alignment.CenterVertically) {
                val avatarPath = user?.avatarPath
                if (avatarPath != null && File(avatarPath).exists()) {
                    AsyncImage(
                        File(avatarPath),
                        null,
                        Modifier
                            .size(72.dp)
                            .clip(CircleShape),
                        contentScale = ContentScale.Crop,
                    )
                } else {
                    Box(
                        Modifier
                            .size(72.dp)
                            .clip(CircleShape)
                            .background(MaterialTheme.colorScheme.primaryContainer),
                        contentAlignment = Alignment.Center,
                    ) {
                        Icon(
                            Icons.Default.Person,
                            null,
                            Modifier.size(36.dp),
                            tint = MaterialTheme.colorScheme.onPrimaryContainer,
                        )
                    }
                }
                Spacer(Modifier.width(16.dp))
                Column {
                    Text(user?.nickname.orEmpty(), style = MaterialTheme.typography.titleLarge)
                    Text("学号：${user?.studentId.orEmpty()}", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    user?.phone?.let { if (it.isNotBlank()) Text("电话：$it", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant) }
                }
            }

            // 编辑资料
            OutlinedButton(
                onClick = { nav.navigate("editProfile") },
                Modifier.fillMaxWidth(),
            ) {
                Icon(Icons.Default.Edit, null, Modifier.size(18.dp))
                Spacer(Modifier.width(8.dp))
                Text("编辑资料")
            }

            Spacer(Modifier.height(4.dp))

            // 我的订单
            Button(onClick = { nav.navigate("orders") }, Modifier.fillMaxWidth()) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(Icons.Default.ReceiptLong, contentDescription = null)
                    Spacer(Modifier.width(8.dp))
                    Text("我的订单")
                }
            }

            // 我的发布
            Button(onClick = { nav.navigate("myPublished") }, Modifier.fillMaxWidth()) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(Icons.Default.Storefront, contentDescription = null)
                    Spacer(Modifier.width(8.dp))
                    Text("我的发布")
                }
            }

            // 通知
            OutlinedButton(
                onClick = { nav.navigate("notifications") },
                Modifier.fillMaxWidth(),
            ) {
                BadgedBox(
                    badge = {
                        if (unreadNotif > 0) {
                            Badge { Text(if (unreadNotif > 99) "99+" else unreadNotif.toString()) }
                        }
                    },
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Icon(Icons.Default.Notifications, contentDescription = null)
                        Spacer(Modifier.width(8.dp))
                        Text("通知")
                    }
                }
            }

            Spacer(Modifier.weight(1f))

            // 退出登录
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

// ─────────────────────────────── 我的发布 ───────────────────────────────

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MyPublishedScreen(nav: NavHostController, userId: Long) {
    val c = LocalAppContainer.current
    val items by c.products.observeBySeller(userId).collectAsStateWithLifecycle(initialValue = emptyList())
    Scaffold(
        topBar = {
            CenterAlignedTopAppBar(
                title = { Text("我的发布") },
                navigationIcon = {
                    IconButton(onClick = { nav.popBackStack() }) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, "back")
                    }
                },
            )
        },
    ) { padding ->
        if (items.isEmpty()) {
            Box(Modifier.padding(padding).fillMaxSize(), contentAlignment = Alignment.Center) {
                Text("还没有发布商品", color = MaterialTheme.colorScheme.onSurfaceVariant)
            }
            return@Scaffold
        }
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
                    colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant),
                ) {
                    Row(Modifier.padding(12.dp)) {
                        val path = p.imageLocalPath
                        if (path != null && File(path).exists()) {
                            AsyncImage(
                                File(path), null,
                                Modifier.size(72.dp).clip(MaterialTheme.shapes.small),
                                contentScale = ContentScale.Crop,
                            )
                        } else {
                            Spacer(Modifier.size(72.dp))
                        }
                        Spacer(Modifier.width(12.dp))
                        Column(Modifier.weight(1f)) {
                            Text(p.title, style = MaterialTheme.typography.titleSmall, maxLines = 1, overflow = TextOverflow.Ellipsis)
                            Text("¥${"%.2f".format(p.priceCents / 100.0)}", color = MaterialTheme.colorScheme.primary)
                            Text(
                                if (p.status == ProductStatus.ON_SALE) "在售" else "已售",
                                style = MaterialTheme.typography.bodySmall,
                                color = if (p.status == ProductStatus.ON_SALE) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.error,
                            )
                        }
                    }
                }
            }
        }
    }
}

// ─────────────────────────────── 通知列表 ───────────────────────────────

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NotificationScreen(nav: NavHostController, userId: Long) {
    val c = LocalAppContainer.current
    val scope = rememberCoroutineScope()
    val notifications by c.db.notificationDao().observeByUser(userId)
        .collectAsStateWithLifecycle(initialValue = emptyList())

    // 进入时标记全部已读
    LaunchedEffect(Unit) {
        c.db.notificationDao().markAllRead(userId)
    }

    Scaffold(
        topBar = {
            CenterAlignedTopAppBar(
                title = { Text("通知") },
                navigationIcon = {
                    IconButton(onClick = { nav.popBackStack() }) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, "back")
                    }
                },
            )
        },
    ) { padding ->
        if (notifications.isEmpty()) {
            Box(Modifier.padding(padding).fillMaxSize(), contentAlignment = Alignment.Center) {
                Text("暂无通知", color = MaterialTheme.colorScheme.onSurfaceVariant)
            }
            return@Scaffold
        }
        LazyColumn(
            Modifier.padding(padding).fillMaxSize(),
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            items(notifications, key = { it.id }) { n ->
                Card(
                    Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(
                        containerColor = if (n.isRead == 0)
                            MaterialTheme.colorScheme.surfaceVariant
                        else
                            MaterialTheme.colorScheme.surface,
                    ),
                ) {
                    Column(Modifier.padding(12.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Text(
                                NotificationType.label(n.type),
                                style = MaterialTheme.typography.titleSmall,
                                modifier = Modifier.weight(1f),
                            )
                            if (n.isRead == 0) {
                                Badge(containerColor = MaterialTheme.colorScheme.error) { Text("新") }
                            }
                        }
                        Text(n.title, style = MaterialTheme.typography.bodyMedium)
                        Text(n.content, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                        Text(formatTime(n.createdAt), style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                }
            }
        }
    }
}

// ─────────────────────────────── 编辑资料 ───────────────────────────────

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun EditProfileScreen(nav: NavHostController, userId: Long) {
    val c = LocalAppContainer.current
    val scope = rememberCoroutineScope()
    val user by c.db.userDao().observeUser(userId).collectAsStateWithLifecycle(initialValue = null)

    var nickname by remember { mutableStateOf("") }
    var phone by remember { mutableStateOf("") }
    var avatarUri by remember { mutableStateOf<Uri?>(null) }
    var err by remember { mutableStateOf<String?>(null) }
    var ok by remember { mutableStateOf<String?>(null) }

    // 用户数据加载后填入表单
    LaunchedEffect(user) {
        user?.let {
            if (nickname.isEmpty()) nickname = it.nickname
            if (phone.isEmpty()) phone = it.phone.orEmpty()
        }
    }

    val pickAvatar = rememberLauncherForActivityResult(PickVisualMedia()) { avatarUri = it }

    Scaffold(
        topBar = {
            CenterAlignedTopAppBar(
                title = { Text("编辑资料") },
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
                .padding(16.dp)
                .fillMaxSize()
                .verticalScroll(rememberScrollState()),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            // 头像预览
            Box(contentAlignment = Alignment.BottomEnd) {
                val displayPath = avatarUri?.toString() ?: user?.avatarPath
                if (displayPath != null && (avatarUri != null || (user?.avatarPath != null && File(user!!.avatarPath!!).exists()))) {
                    AsyncImage(
                        model = avatarUri ?: File(user!!.avatarPath!!),
                        contentDescription = null,
                        Modifier
                            .size(96.dp)
                            .clip(CircleShape),
                        contentScale = ContentScale.Crop,
                    )
                } else {
                    Box(
                        Modifier
                            .size(96.dp)
                            .clip(CircleShape)
                            .background(MaterialTheme.colorScheme.primaryContainer),
                        contentAlignment = Alignment.Center,
                    ) {
                        Icon(
                            Icons.Default.Person, null, Modifier.size(48.dp),
                            tint = MaterialTheme.colorScheme.onPrimaryContainer,
                        )
                    }
                }
                IconButton(
                    onClick = {
                        pickAvatar.launch(PickVisualMediaRequest(PickVisualMedia.ImageOnly))
                    },
                    Modifier.size(32.dp),
                ) {
                    Icon(Icons.Default.Edit, "更换头像", Modifier.size(16.dp))
                }
            }

            OutlinedTextField(
                nickname,
                { nickname = it },
                Modifier.fillMaxWidth(),
                label = { Text("昵称") },
                singleLine = true,
            )
            OutlinedTextField(
                phone,
                { phone = it },
                Modifier.fillMaxWidth(),
                label = { Text("电话（可选）") },
                singleLine = true,
            )

            err?.let { Text(it, color = MaterialTheme.colorScheme.error) }
            ok?.let { Text(it, color = MaterialTheme.colorScheme.primary) }

            Button(
                onClick = {
                    scope.launch {
                        err = null
                        ok = null
                        if (nickname.isBlank()) {
                            err = "昵称不能为空"
                            return@launch
                        }
                        // 更新头像
                        avatarUri?.let { uri ->
                            val avatarDir = File(c.appContext.filesDir, "avatars").apply { mkdirs() }
                            val outFile = File(avatarDir, "user_${userId}.jpg")
                            c.appContext.contentResolver.openInputStream(uri)?.use { input ->
                                outFile.outputStream().use { output -> input.copyTo(output) }
                            }
                            c.auth.updateAvatar(userId, outFile.absolutePath)
                        }
                        c.auth.updateProfile(userId, nickname.trim(), phone.ifBlank { null })
                        ok = "保存成功"
                    }
                },
                Modifier.fillMaxWidth(),
            ) { Text("保存") }
        }
    }
}

