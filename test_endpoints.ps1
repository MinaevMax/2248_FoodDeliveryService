#!/usr/bin/env pwsh

# Тестирование API эндпоинтов Food Delivery Service

$BASE_URL = "http://localhost:8080"
$ContentType = "application/json"

Write-Host "=== Testing Food Delivery Service API ===" -ForegroundColor Yellow
Write-Host ""

# Test 1: Register user
Write-Host "1. Testing POST /auth/register" -ForegroundColor Yellow
$registerBody = @{
    login = "testuser"
    password = "password123"
} | ConvertTo-Json

try {
    $registerResponse = Invoke-WebRequest -Uri "$BASE_URL/auth/register" `
        -Method POST `
        -Headers @{"Content-Type" = $ContentType} `
        -Body $registerBody
    
    Write-Host "Response: $($registerResponse.Content)" -ForegroundColor Green
    Write-Host "✓ Registration successful" -ForegroundColor Green
    
    $userData = $registerResponse.Content | ConvertFrom-Json
    $userId = $userData.id
}
catch {
    Write-Host "✗ Registration failed: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    Write-Host "$($_.Exception.Response | Get-Member)" -ForegroundColor Red
    $userId = $null
}

Write-Host ""

# Test 2: Login user
Write-Host "2. Testing POST /auth/login" -ForegroundColor Yellow
$loginBody = @{
    login = "testuser"
    password = "password123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-WebRequest -Uri "$BASE_URL/auth/login" `
        -Method POST `
        -Headers @{"Content-Type" = $ContentType} `
        -Body $loginBody
    
    Write-Host "Response: $($loginResponse.Content)" -ForegroundColor Green
    
    $loginData = $loginResponse.Content | ConvertFrom-Json
    $token = $loginData.token
    
    Write-Host "✓ Login successful, token: $($token.Substring(0, 20))..." -ForegroundColor Green
}
catch {
    Write-Host "✗ Login failed: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    $token = $null
}

Write-Host ""

# Test 3: Create order
if ($token) {
    Write-Host "3. Testing POST /orders/create (with token)" -ForegroundColor Yellow
    $orderBody = @{
        amount = 100
    } | ConvertTo-Json
    
    try {
        $createOrderResponse = Invoke-WebRequest -Uri "$BASE_URL/orders/create" `
            -Method POST `
            -Headers @{
                "Authorization" = "Bearer $token"
                "Content-Type" = $ContentType
            } `
            -Body $orderBody
        
        Write-Host "Response: $($createOrderResponse.Content)" -ForegroundColor Green
        Write-Host "✓ Order created successfully" -ForegroundColor Green
    }
    catch {
        Write-Host "✗ Order creation failed: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
        Write-Host "$($_.Exception.Message)" -ForegroundColor Red
    }
    
    Write-Host ""
    
    # Test 4: Get orders list
    Write-Host "4. Testing GET /orders/list (with token)" -ForegroundColor Yellow
    
    try {
        $getOrdersResponse = Invoke-WebRequest -Uri "$BASE_URL/orders/list?active=true" `
            -Method GET `
            -Headers @{
                "Authorization" = "Bearer $token"
                "Content-Type" = $ContentType
            }
        
        Write-Host "Response: $($getOrdersResponse.Content)" -ForegroundColor Green
        Write-Host "✓ Orders list retrieved successfully" -ForegroundColor Green
    }
    catch {
        Write-Host "✗ Orders list retrieval failed: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    }
    
    Write-Host ""
}

# Test 5: Test without token
Write-Host "5. Testing /orders/create (without token - should fail)" -ForegroundColor Yellow
$orderBody = @{
    amount = 100
} | ConvertTo-Json

try {
    $noTokenResponse = Invoke-WebRequest -Uri "$BASE_URL/orders/create" `
        -Method POST `
        -Headers @{"Content-Type" = $ContentType} `
        -Body $orderBody
    
    Write-Host "✗ Should have rejected request without token" -ForegroundColor Red
}
catch {
    if ($_.Exception.Response.StatusCode -eq 401) {
        Write-Host "Response: $($_.Exception.Response.StatusCode) Unauthorized" -ForegroundColor Green
        Write-Host "✓ Correctly rejected request without token" -ForegroundColor Green
    }
    else {
        Write-Host "✗ Unexpected error: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "=== Tests Complete ===" -ForegroundColor Yellow
