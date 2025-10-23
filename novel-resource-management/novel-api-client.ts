/**
 * Novel API Client - 用于Next.js项目中调用后端接口
 * 支持加密和非加密两种请求方式
 */

// 基础响应类型
export interface ApiResponse<T = any> {
    success: boolean;
    data?: T;
    error?: string;
    message?: string;
}

// 小说数据类型
export interface NovelData {
    id: string;
    author: string;
    storyOutline: string;
    subsections: string;
    characters: string;
    items: string;
    totalScenes: string;
    createdAt?: string;
    updatedAt?: string;
}

// 用户积分数据类型
export interface UserCreditData {
    userId: string;
    credit: number;
    totalUsed: number;
    totalRecharge: number;
    createdAt?: string;
    updatedAt?: string;
}

// 加密请求类型
export interface EncryptedRequest {
    encryptedData: string;
}

class NovelAPIClient {
    private baseURL: string;
    private useRSA: boolean;
    private defaultHeaders: Record<string, string>;

    constructor(
        baseURL: string = 'http://localhost:8080',
        useRSA: boolean = false
    ) {
        this.baseURL = baseURL;
        this.useRSA = useRSA;
        this.defaultHeaders = {
            'Content-Type': 'application/json',
        };
    }

    /**
     * 设置是否使用RSA加密
     */
    setUseRSA(useRSA: boolean): void {
        this.useRSA = useRSA;
    }

    /**
     * 设置基础URL
     */
    setBaseURL(baseURL: string): void {
        this.baseURL = baseURL;
    }

    /**
     * 通用请求方法
     */
    private async request<T = any>(
        method: string,
        endpoint: string,
        data?: any,
        options?: RequestInit
    ): Promise<ApiResponse<T>> {
        const url = `${this.baseURL}${endpoint}`;

        const config: RequestInit = {
            method,
            headers: { ...this.defaultHeaders },
            ...options,
        };

        // 处理请求体
        if (data) {
            if (this.useRSA && ['POST', 'PUT'].includes(method)) {
                // 使用RSA加密
                const encryptedData = await this.encryptData(data);
                config.body = JSON.stringify({
                    encryptedData: encryptedData,
                });
                config.headers!['X-Encrypted-Request'] = 'true';
            } else {
                // 普通请求
                config.body = JSON.stringify(data);
            }
        }

        try {
            const response = await fetch(url, config);
            const responseData = await response.json().catch(() => ({}));

            return {
                success: response.ok,
                data: response.ok ? responseData : undefined,
                error: response.ok
                    ? undefined
                    : responseData.error || `HTTP ${response.status}`,
                message: responseData.message,
            };
        } catch (error) {
            return {
                success: false,
                error: error instanceof Error ? error.message : 'Network error',
            };
        }
    }

    /**
     * RSA加密数据 (重要提示：这是一个占位符函数)
     * 在实际项目中，您必须实现真正的RSA加密逻辑！
     */
    private async encryptData(data: any): Promise<string> {
        // ⚠️ 警告：这不是真正的RSA加密！
        // 这只是一个Base64编码，用于测试API调用流程
        // 实际使用时，您必须替换为真正的RSA加密实现

        const jsonData = JSON.stringify(data);

        // TODO: 实现真正的RSA加密
        // 方案1: 使用Web Crypto API
        // 方案2: 使用node-rsa库（Node.js环境）
        // 方案3: 使用jsencrypt库（浏览器环境）
        // 方案4: 调用专门的加密接口

        console.warn(
            '⚠️ 当前使用的是Base64编码，不是真正的RSA加密！请实现真正的加密逻辑！'
        );

        return btoa(jsonData);
    }

    // ========== 小说相关接口 ==========

    /**
     * 获取所有小说
     */
    async getAllNovels(): Promise<ApiResponse<NovelData[]>> {
        return this.request<NovelData[]>('GET', '/api/v1/novels');
    }

    /**
     * 根据ID获取小说
     */
    async getNovel(id: string): Promise<ApiResponse<NovelData>> {
        return this.request<NovelData>('GET', `/api/v1/novels/${id}`);
    }

    /**
     * 创建小说
     */
    async createNovel(
        novel: Omit<NovelData, 'createdAt' | 'updatedAt'>
    ): Promise<ApiResponse<{ message: string; id: string }>> {
        return this.request('POST', '/api/v1/novels', novel);
    }

    /**
     * 更新小说
     */
    async updateNovel(
        id: string,
        novel: Partial<NovelData>
    ): Promise<ApiResponse<{ message: string; id: string }>> {
        return this.request('PUT', `/api/v1/novels/${id}`, novel);
    }

    /**
     * 删除小说
     */
    async deleteNovel(
        id: string
    ): Promise<ApiResponse<{ message: string; id: string }>> {
        return this.request('DELETE', `/api/v1/novels/${id}`);
    }

    // ========== 用户积分相关接口 ==========

    /**
     * 获取所有用户积分
     */
    async getAllUserCredits(): Promise<ApiResponse<UserCreditData[]>> {
        return this.request<UserCreditData[]>('GET', '/api/v1/users');
    }

    /**
     * 根据ID获取用户积分
     */
    async getUserCredit(id: string): Promise<ApiResponse<UserCreditData>> {
        return this.request<UserCreditData>('GET', `/api/v1/users/${id}`);
    }

    /**
     * 创建用户积分
     */
    async createUserCredit(
        credit: Omit<UserCreditData, 'createdAt' | 'updatedAt'>
    ): Promise<ApiResponse<{ message: string; id: string }>> {
        return this.request('POST', '/api/v1/users', credit);
    }

    /**
     * 更新用户积分
     */
    async updateUserCredit(
        id: string,
        credit: Partial<UserCreditData>
    ): Promise<ApiResponse<{ message: string; id: string }>> {
        return this.request('PUT', `/api/v1/users/${id}`, credit);
    }

    /**
     * 删除用户积分
     */
    async deleteUserCredit(
        id: string
    ): Promise<ApiResponse<{ message: string; id: string }>> {
        return this.request('DELETE', `/api/v1/users/${id}`);
    }

    // ========== 系统相关接口 ==========

    /**
     * 健康检查
     */
    async healthCheck(): Promise<
        ApiResponse<{ status: string; message: string; time: string }>
    > {
        return this.request('GET', '/health');
    }

    /**
     * 流式事件监听
     */
    async streamEvents(): Promise<EventSource> {
        const url = `${this.baseURL}/api/v1/events/listen`;
        return new EventSource(url);
    }
}

// ========== 工厂函数和便捷方法 ==========

/**
 * 创建普通API客户端（不加密）
 */
export function createAPIClient(baseURL?: string): NovelAPIClient {
    return new NovelAPIClient(baseURL, false);
}

/**
 * 创建加密API客户端（使用RSA加密）
 */
export function createSecureAPIClient(baseURL?: string): NovelAPIClient {
    return new NovelAPIClient(baseURL, true);
}

/**
 * 创建API客户端（可配置加密选项）
 */
export function createConfigurableAPIClient(
    baseURL?: string,
    useRSA?: boolean
): NovelAPIClient {
    return new NovelAPIClient(baseURL, useRSA);
}

// ========== React Hooks ==========

import { useState, useEffect } from 'react';

/**
 * 使用小说数据的Hook
 */
export function useNovels(apiClient: NovelAPIClient) {
    const [novels, setNovels] = useState<NovelData[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const fetchNovels = async () => {
        setLoading(true);
        setError(null);

        try {
            const response = await apiClient.getAllNovels();
            if (response.success && response.data) {
                setNovels(response.data);
            } else {
                setError(response.error || '获取小说列表失败');
            }
        } catch (err) {
            setError(err instanceof Error ? err.message : '未知错误');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchNovels();
    }, [apiClient]);

    return {
        novels,
        loading,
        error,
        refetch: fetchNovels,
    };
}

/**
 * 使用单个小说数据的Hook
 */
export function useNovel(apiClient: NovelAPIClient, id: string) {
    const [novel, setNovel] = useState<NovelData | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const fetchNovel = async () => {
        if (!id) return;

        setLoading(true);
        setError(null);

        try {
            const response = await apiClient.getNovel(id);
            if (response.success && response.data) {
                setNovel(response.data);
            } else {
                setError(response.error || '获取小说详情失败');
            }
        } catch (err) {
            setError(err instanceof Error ? err.message : '未知错误');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchNovel();
    }, [apiClient, id]);

    return {
        novel,
        loading,
        error,
        refetch: fetchNovel,
    };
}

export default NovelAPIClient;
