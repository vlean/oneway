// 全局共享数据示例
import { DEFAULT_NAME } from '@/constants';
import { useState } from 'react';

const useUser = () => {
  const [email, setEmail] = useState<string>("");
  return {
    email,
    setEmail,
  };
};

export default useUser;
