import { message } from 'antd';

export async function fetchRequest(url: string, options: RequestInit = {}) {
  const defaultOptions: RequestInit = {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const authHeaders: Record<string, string> = {}
  const token = localStorage.getItem('token')
  if (token) {
    authHeaders.Authorization = `Bearer ${token}`
  }

  const mergedOptions: RequestInit = {
    ...defaultOptions,
    ...options,
    headers: {
      ...defaultOptions.headers,
      ...options.headers,
      ...authHeaders,
    },
  };

  if (mergedOptions.body && typeof mergedOptions.body === 'object') {
    mergedOptions.body = JSON.stringify(mergedOptions.body);
  }

  try {
    const response = await fetch(url, mergedOptions);
    if (response.status === 401) {
      message.error('请先登录');
      window.location.href = '/login';
    }
    if (response.status === 403) {
      throw new Error('权限不足');
    }
    if (response.status === 404) {
      throw new Error('接口不存在');
    }
    
    const data = await response.json();
    if (!response.ok) {
      throw new Error(data.error || `请求失败，状态码：${response.status}`);
    }

    return data;
  } catch (error: any) {
    message.error(error.message);
    throw error;
  }
}