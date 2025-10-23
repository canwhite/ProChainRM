/**
 * Next.js中使用Novel API客户端的示例
 * 这个文件展示了如何在Next.js项目中使用API客户端
 */

import { 
  createAPIClient, 
  createSecureAPIClient, 
  createConfigurableAPIClient,
  useNovels,
  useNovel,
  NovelData,
  UserCreditData
} from './novel-api-client';

// ========== 基本使用示例 ==========

// 1. 创建API客户端实例
const apiClient = createAPIClient('http://localhost:8080'); // 普通请求
const secureApiClient = createSecureAPIClient('http://localhost:8080'); // 加密请求

// 2. 在普通JavaScript/TypeScript中使用
class NovelService {
  private client: any;

  constructor(useSecure = false) {
    this.client = useSecure ? createSecureAPIClient() : createAPIClient();
  }

  // 获取所有小说
  async getAllNovels() {
    const response = await this.client.getAllNovels();
    if (response.success) {
      console.log('小说列表:', response.data);
      return response.data;
    } else {
      console.error('获取失败:', response.error);
      throw new Error(response.error);
    }
  }

  // 创建小说
  async createNovel(novelData: Omit<NovelData, 'createdAt' | 'updatedAt'>) {
    const response = await this.client.createNovel(novelData);
    if (response.success) {
      console.log('创建成功:', response.data);
      return response.data;
    } else {
      console.error('创建失败:', response.error);
      throw new Error(response.error);
    }
  }

  // 更新小说
  async updateNovel(id: string, novelData: Partial<NovelData>) {
    const response = await this.client.updateNovel(id, novelData);
    if (response.success) {
      console.log('更新成功:', response.data);
      return response.data;
    } else {
      console.error('更新失败:', response.error);
      throw new Error(response.error);
    }
  }

  // 删除小说
  async deleteNovel(id: string) {
    const response = await this.client.deleteNovel(id);
    if (response.success) {
      console.log('删除成功:', response.data);
      return response.data;
    } else {
      console.error('删除失败:', response.error);
      throw new Error(response.error);
    }
  }
}

// ========== Next.js页面组件示例 ==========

/*
// 在Next.js页面中使用示例：

import { useState, useEffect } from 'react';
import { createAPIClient, useNovels } from '../path/to/novel-api-client';

export default function NovelsPage() {
  // 使用Hook获取小说列表
  const { novels, loading, error, refetch } = useNovels(createAPIClient());
  
  // 或者手动管理状态
  const [novels, setNovels] = useState<NovelData[]>([]);
  const [loading, setLoading] = useState(false);
  
  const client = createAPIClient();

  useEffect(() => {
    loadNovels();
  }, []);

  const loadNovels = async () => {
    setLoading(true);
    try {
      const response = await client.getAllNovels();
      if (response.success) {
        setNovels(response.data || []);
      }
    } catch (error) {
      console.error('加载失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateNovel = async () => {
    const newNovel: Omit<NovelData, 'createdAt' | 'updatedAt'> = {
      id: `novel-${Date.now()}`,
      author: '新作者',
      storyOutline: '新故事大纲',
      subsections: '章节1,章节2',
      characters: '主角,配角',
      items: '道具1,道具2',
      totalScenes: '5'
    };

    try {
      const response = await client.createNovel(newNovel);
      if (response.success) {
        alert('创建成功！');
        loadNovels(); // 重新加载列表
      } else {
        alert('创建失败: ' + response.error);
      }
    } catch (error) {
      alert('创建失败: ' + error);
    }
  };

  if (loading) return <div>加载中...</div>;
  if (error) return <div>错误: {error}</div>;

  return (
    <div>
      <h1>小说管理</h1>
      <button onClick={handleCreateNovel}>创建新小说</button>
      
      <ul>
        {novels.map(novel => (
          <li key={novel.id}>
            <h3>{novel.author} - {novel.id}</h3>
            <p>{novel.storyOutline}</p>
          </li>
        ))}
      </ul>
    </div>
  );
}
*/

// ========== Next.js API路由示例 ==========

/*
// 在pages/api/novels.ts中处理客户端请求：

import type { NextApiRequest, NextApiResponse } from 'next';
import { createConfigurableAPIClient } from '../../path/to/novel-api-client';

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  const client = createConfigurableAPIClient(
    process.env.API_BASE_URL || 'http://localhost:8080',
    process.env.USE_RSA_ENCRYPTION === 'true'
  );

  try {
    switch (req.method) {
      case 'GET':
        const novelsResponse = await client.getAllNovels();
        return res.status(novelsResponse.success ? 200 : 400).json(novelsResponse);
      
      case 'POST':
        const createResponse = await client.createNovel(req.body);
        return res.status(createResponse.success ? 200 : 400).json(createResponse);
      
      default:
        return res.status(405).json({ success: false, error: 'Method not allowed' });
    }
  } catch (error) {
    return res.status(500).json({ 
      success: false, 
      error: error instanceof Error ? error.message : 'Unknown error' 
    });
  }
}
*/

// ========== 服务端渲染示例 ==========

/*
// 在getServerSideProps中使用：

import { GetServerSideProps } from 'next';
import { createAPIClient } from '../path/to/novel-api-client';

export const getServerSideProps: GetServerSideProps = async (context) => {
  const client = createAPIClient(process.env.API_BASE_URL);
  
  try {
    const response = await client.getAllNovels();
    
    return {
      props: {
        novels: response.success ? response.data : [],
        error: response.success ? null : response.error,
      },
    };
  } catch (error) {
    return {
      props: {
        novels: [],
        error: error instanceof Error ? error.message : 'Unknown error',
      },
    };
  }
};
*/

// ========== 环境变量配置 ==========

/*
// 在.env.local文件中配置：

API_BASE_URL=http://localhost:8080
USE_RSA_ENCRYPTION=true

// 在.env.production文件中配置生产环境：

API_BASE_URL=https://your-api-domain.com
USE_RSA_ENCRYPTION=true
*/

// ========== RSA加密实现建议 ==========

/*
// 如果需要真正的RSA加密，可以创建一个加密工具类：

// lib/rsa.ts
export class RSAEncryption {
  private publicKey: string;

  constructor(publicKey: string) {
    this.publicKey = publicKey;
  }

  async encrypt(data: string): Promise<string> {
    // 使用Web Crypto API进行RSA加密
    // 这里需要根据您的实际RSA公钥格式来实现
    const encoder = new TextEncoder();
    const dataBuffer = encoder.encode(data);
    
    // 导入公钥
    const publicKey = await window.crypto.subtle.importKey(
      'spki',
      this.base64ToArrayBuffer(this.publicKey),
      { name: 'RSA-OAEP', hash: 'SHA-256' },
      false,
      ['encrypt']
    );

    // 加密数据
    const encrypted = await window.crypto.subtle.encrypt(
      { name: 'RSA-OAEP' },
      publicKey,
      dataBuffer
    );

    return this.arrayBufferToBase64(encrypted);
  }

  private base64ToArrayBuffer(base64: string): ArrayBuffer {
    const binaryString = atob(base64);
    const bytes = new Uint8Array(binaryString.length);
    for (let i = 0; i < binaryString.length; i++) {
      bytes[i] = binaryString.charCodeAt(i);
    }
    return bytes.buffer;
  }

  private arrayBufferToBase64(buffer: ArrayBuffer): string {
    const bytes = new Uint8Array(buffer);
    let binary = '';
    for (let i = 0; i < bytes.byteLength; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary);
  }
}

// 然后修改NovelAPIClient中的encryptData方法：
private async encryptData(data: any): Promise<string> {
  const rsa = new RSAEncryption(this.publicKey);
  const jsonData = JSON.stringify(data);
  return rsa.encrypt(jsonData);
}
*/

// ========== 使用示例 ==========

// 创建使用示例
const exampleUsage = async () => {
  // 创建客户端
  const client = createConfigurableAPIClient('http://localhost:8080', true);
  
  // 健康检查
  const health = await client.healthCheck();
  console.log('健康检查:', health);
  
  // 获取小说列表
  const novels = await client.getAllNovels();
  console.log('小说列表:', novels);
  
  // 创建小说
  const newNovel = await client.createNovel({
    id: 'test-novel-001',
    author: '测试作者',
    storyOutline: '这是一个测试故事',
    subsections: '章节1,章节2',
    characters: '主角,配角',
    items: '道具1,道具2',
    totalScenes: '5'
  });
  console.log('创建结果:', newNovel);
  
  // 获取用户积分
  const credits = await client.getAllUserCredits();
  console.log('用户积分:', credits);
};

export { NovelService };