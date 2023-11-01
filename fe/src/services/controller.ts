import { request } from '@umijs/max';

export async function forwardList(
  params: {
    keyword?: string;
    page?: number;
    page_size?: number;
  },
  options?: { [key: string]: any },
) {
  return request<API.Response>('/api/forward', {
    method: 'GET',
    params: {
        page: 1,
        page_size: 20,
      ...params,
    },
    ...(options || {}),
  });
}
export async function forwardSave(data: any) {
    return request<API.Response>('/api/forward/save', {
        method: 'POST',
        data: data, 
    })
}
export async function forwardDel(data: any) {
    return request<API.Response>('/api/forward/delete', {
        method: 'POST',
        data: data, 
    })
}

export async function userinfo() {
    return request<API.Response>('/api/user', {
        method: 'GET',
    })
}


export async function systemConfig() {
  return request<API.Response>('/api/system/config', {
      method: 'GET',
  })
}

export async function systemConfigUpdate(data: any) {
  return request<API.Response>('/api/system/config/update', {
      method: 'POST',
      data: data, 
  })
}
