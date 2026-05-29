# API 设计

## 基础信息

- **Base URL**: `http://localhost:8080`
- **Content-Type**: `application/json`

## 接口列表

### 健康检查

```
GET /api/health
```

**响应**
```json
{
  "status": "ok"
}
```

### 图片管理

待补充 - 根据具体需求设计接口

#### 获取图片列表

```
GET /api/images
```

#### 获取单张图片

```
GET /api/images/:id
```

#### 上传图片

```
POST /api/images
Content-Type: multipart/form-data
```

#### 删除图片

```
DELETE /api/images/:id
```

## 错误码

| Code | 说明 |
|------|------|
| 400  | 请求参数错误 |
| 404  | 资源不存在 |
| 500  | 服务器错误 |
