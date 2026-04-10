/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useContext, useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  Button,
  Input,
  ScrollItem,
  ScrollList,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import { IconCopy, IconFile, IconPlay, IconRefresh } from '@douyinfe/semi-icons';
import { marked } from 'marked';
import {
  AzureAI,
  Claude,
  DeepSeek,
  Gemini,
  Hunyuan,
  OpenAI,
  Qwen,
  XAI,
} from '@lobehub/icons';
import { API, copy, showError, showSuccess } from '../../helpers';
import { API_ENDPOINTS } from '../../constants/common.constant';
import { StatusContext } from '../../context/Status';
import { useActualTheme } from '../../context/Theme';
import { useIsMobile } from '../../hooks/common/useIsMobile';
import { useTranslation } from 'react-i18next';
import NoticeModal from '../../components/layout/NoticeModal';

const { Text } = Typography;

const packagePlans = [
  {
    id: 'promo-10',
    name: '限时特惠 · 入门订阅',
    price: '10 元',
    token: '1000W Token',
    group: '限时特惠分组',
    desc: '适合个人开发者、轻量业务与测试场景，快速开通即刻可用。',
    perks: ['主流模型统一接口', '稳定中转链路', '专属高峰期调度'],
  },
  {
    id: 'promo-50',
    name: '限时特惠 · 进阶订阅',
    price: '50 元',
    token: '7000W Token',
    group: '限时特惠分组',
    desc: '适合团队与生产业务，提供更高吞吐、更低抖动与更优成本结构。',
    perks: ['高并发优先通道', '更强容量弹性', '专业级运营保障'],
  },
];

const coreAdvantages = [
  {
    title: '专业中转架构',
    detail:
      '统一适配多家模型厂商，标准 OpenAI 风格协议，切换供应商无需改业务逻辑。',
  },
  {
    title: '高可用调度',
    detail: '多通道健康探测与自动路由，减少单点波动影响，保障调用连续性。',
  },
  {
    title: '成本与性能平衡',
    detail:
      '按需选择策略、分组与套餐，在稳定与成本之间持续优化你的业务曲线。',
  },
];

const providerLogos = [
  { name: 'OpenAI', Comp: OpenAI },
  { name: 'Claude', Comp: Claude.Color },
  { name: 'Gemini', Comp: Gemini.Color },
  { name: 'DeepSeek', Comp: DeepSeek.Color },
  { name: 'Qwen', Comp: Qwen.Color },
  { name: 'xAI', Comp: XAI },
  { name: 'AzureAI', Comp: AzureAI.Color },
  { name: 'Hunyuan', Comp: Hunyuan.Color },
];

const connectivityTargets = [
  { key: 'status', name: '系统状态接口', path: '/api/status' },
  { key: 'models', name: '模型列表接口', path: '/v1/models' },
  { key: 'homepage', name: '首页内容接口', path: '/api/home_page_content' },
  { key: 'notice', name: '公告接口', path: '/api/notice' },
];

const Home = () => {
  const { i18n } = useTranslation();
  const [statusState] = useContext(StatusContext);
  const actualTheme = useActualTheme();
  const isMobile = useIsMobile();

  const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
  const [homePageContent, setHomePageContent] = useState('');
  const [noticeVisible, setNoticeVisible] = useState(false);
  const [endpointIndex, setEndpointIndex] = useState(0);

  const [checkingConnectivity, setCheckingConnectivity] = useState(false);
  const [connectivitySummary, setConnectivitySummary] = useState(
    '点击按钮检查当前站点模型连通率',
  );
  const [connectivityRate, setConnectivityRate] = useState(null);
  const [connectivityDetails, setConnectivityDetails] = useState(
    connectivityTargets.map((item) => ({
      ...item,
      ok: null,
      status: null,
      latency: null,
      reason: '待检测',
    })),
  );

  const docsLink = statusState?.status?.docs_link || '';
  const serverAddress =
    statusState?.status?.server_address || `${window.location.origin}`;
  const endpointItems = useMemo(
    () => API_ENDPOINTS.map((endpoint) => ({ value: endpoint })),
    [],
  );

  const displayHomePageContent = async () => {
    setHomePageContent(localStorage.getItem('home_page_content') || '');
    const res = await API.get('/api/home_page_content');
    const { success, message, data } = res.data;
    if (success) {
      let content = data;
      if (!data.startsWith('https://')) {
        content = marked.parse(data);
      }
      setHomePageContent(content);
      localStorage.setItem('home_page_content', content);

      if (data.startsWith('https://')) {
        const iframe = document.querySelector('iframe');
        if (iframe) {
          iframe.onload = () => {
            iframe.contentWindow.postMessage({ themeMode: actualTheme }, '*');
            iframe.contentWindow.postMessage({ lang: i18n.language }, '*');
          };
        }
      }
    } else {
      showError(message);
      setHomePageContent('主页内容加载失败，请稍后重试。');
    }
    setHomePageContentLoaded(true);
  };

  const handleCopyBaseURL = async () => {
    const ok = await copy(serverAddress);
    if (ok) {
      showSuccess('已复制到剪切板');
    }
  };

  const probeEndpoint = async (target) => {
    const start = performance.now();
    try {
      const res = await API.get(target.path, {
        skipErrorHandler: true,
        validateStatus: () => true,
      });
      const latency = Math.round(performance.now() - start);
      const ok = res.status >= 200 && res.status < 500;
      return {
        ...target,
        ok,
        status: res.status,
        latency,
        reason: ok ? '连通' : `状态码 ${res.status}`,
      };
    } catch (err) {
      return {
        ...target,
        ok: false,
        status: null,
        latency: null,
        reason: err?.code || '请求失败',
      };
    }
  };

  const checkConnectivity = async () => {
    if (checkingConnectivity) {
      return;
    }
    setCheckingConnectivity(true);
    setConnectivitySummary('正在检测接口连通率，请稍候...');

    const results = [];
    for (const target of connectivityTargets) {
      const result = await probeEndpoint(target);
      results.push(result);
      setConnectivityDetails([...results]);
    }

    const successCount = results.filter((item) => item.ok).length;
    const rate = Math.round((successCount / connectivityTargets.length) * 100);
    setConnectivityRate(rate);

    if (rate >= 80) {
      setConnectivitySummary('连通率优秀，当前可稳定承载大模型调用。');
    } else if (rate >= 50) {
      setConnectivitySummary('连通率一般，建议排查部分接口状态。');
    } else {
      setConnectivitySummary('连通率偏低，请检查后端服务与网关配置。');
    }

    setCheckingConnectivity(false);
  };

  useEffect(() => {
    const checkNoticeAndShow = async () => {
      const lastCloseDate = localStorage.getItem('notice_close_date');
      const today = new Date().toDateString();
      if (lastCloseDate !== today) {
        try {
          const res = await API.get('/api/notice');
          const { success, data } = res.data;
          if (success && data && data.trim() !== '') {
            setNoticeVisible(true);
          }
        } catch (error) {
          console.error('获取公告失败:', error);
        }
      }
    };
    checkNoticeAndShow();
  }, []);

  useEffect(() => {
    displayHomePageContent().then();
  }, []);

  useEffect(() => {
    const timer = setInterval(() => {
      setEndpointIndex((prev) => (prev + 1) % endpointItems.length);
    }, 2800);
    return () => clearInterval(timer);
  }, [endpointItems.length]);

  if (!homePageContentLoaded) {
    return (
      <div className='w-full min-h-[70vh] flex items-center justify-center mt-[60px]'>
        <Text type='tertiary'>正在初始化首页...</Text>
      </div>
    );
  }

  if (homePageContent !== '') {
    return (
      <div className='overflow-x-hidden w-full'>
        <NoticeModal
          visible={noticeVisible}
          onClose={() => setNoticeVisible(false)}
          isMobile={isMobile}
        />
        {homePageContent.startsWith('https://') ? (
          <iframe src={homePageContent} className='w-full h-screen border-none' />
        ) : (
          <div
            className='mt-[60px]'
            dangerouslySetInnerHTML={{ __html: homePageContent }}
          />
        )}
      </div>
    );
  }

  return (
    <div className='tech-home w-full overflow-x-hidden'>
      <style>{`
        .tech-home {
          background:
            radial-gradient(circle at 20% 20%, rgba(123, 205, 255, 0.38), transparent 45%),
            radial-gradient(circle at 80% 15%, rgba(105, 171, 255, 0.3), transparent 40%),
            linear-gradient(160deg, #f6fbff 0%, #d8ecff 48%, #cae4ff 100%);
          min-height: calc(100vh - 60px);
          position: relative;
        }
        .tech-grid::before {
          content: "";
          position: absolute;
          inset: 0;
          background-image:
            linear-gradient(rgba(84, 149, 226, 0.15) 1px, transparent 1px),
            linear-gradient(90deg, rgba(84, 149, 226, 0.15) 1px, transparent 1px);
          background-size: 42px 42px;
          mask-image: radial-gradient(circle at center, rgba(0,0,0,1), rgba(0,0,0,0));
          pointer-events: none;
          animation: gridFlow 10s linear infinite;
        }
        .hero-orb {
          position: absolute;
          border-radius: 9999px;
          filter: blur(2px);
          animation: orbFloat 8s ease-in-out infinite;
        }
        .scan-line {
          position: absolute;
          left: 0;
          right: 0;
          height: 2px;
          background: linear-gradient(90deg, transparent, rgba(18, 122, 255, 0.65), transparent);
          animation: scanDown 4.8s linear infinite;
          pointer-events: none;
        }
        .float-particle {
          position: absolute;
          width: 8px;
          height: 8px;
          border-radius: 9999px;
          background: rgba(46, 144, 255, 0.52);
          box-shadow: 0 0 18px rgba(46, 144, 255, 0.65);
          animation: particleRise 7s ease-in-out infinite;
        }
        .tech-card {
          backdrop-filter: blur(8px);
          border: 1px solid rgba(133, 193, 255, 0.6);
          background: linear-gradient(150deg, rgba(255, 255, 255, 0.86), rgba(237, 247, 255, 0.72));
          box-shadow: 0 18px 35px rgba(46, 109, 201, 0.16);
          transition: all 0.28s ease;
        }
        .tech-card:hover {
          transform: translateY(-6px);
          box-shadow: 0 24px 44px rgba(27, 95, 197, 0.22);
          border-color: rgba(60, 150, 255, 0.78);
        }
        .token-glow {
          text-shadow: 0 0 14px rgba(25, 126, 255, 0.36);
        }
        @keyframes orbFloat {
          0%, 100% { transform: translateY(0px) scale(1); }
          50% { transform: translateY(-18px) scale(1.06); }
        }
        @keyframes particleRise {
          0%, 100% { transform: translateY(0px) scale(0.8); opacity: 0.35; }
          50% { transform: translateY(-26px) scale(1); opacity: 0.85; }
        }
        @keyframes scanDown {
          0% { top: 0%; opacity: 0; }
          8% { opacity: 1; }
          92% { opacity: 1; }
          100% { top: 100%; opacity: 0; }
        }
        @keyframes gridFlow {
          0% { transform: translateY(0px); }
          100% { transform: translateY(42px); }
        }
      `}</style>

      <NoticeModal
        visible={noticeVisible}
        onClose={() => setNoticeVisible(false)}
        isMobile={isMobile}
      />

      <div className='mt-[60px] relative tech-grid overflow-hidden'>
        <div className='hero-orb w-[260px] h-[260px] bg-[#7fd3ff]/45 top-12 left-[-80px]' />
        <div
          className='hero-orb w-[200px] h-[200px] bg-[#6fa7ff]/35 top-28 right-[-40px]'
          style={{ animationDelay: '1s' }}
        />
        <div
          className='hero-orb w-[120px] h-[120px] bg-[#7be8ff]/55 bottom-24 left-[22%]'
          style={{ animationDelay: '2s' }}
        />
        <div className='scan-line' />
        {Array.from({ length: 12 }).map((_, idx) => (
          <span
            key={`particle-${idx}`}
            className='float-particle'
            style={{
              left: `${8 + idx * 8}%`,
              top: `${20 + (idx % 5) * 12}%`,
              animationDelay: `${(idx % 4) * 0.7}s`,
            }}
          />
        ))}

        <div className='max-w-7xl mx-auto px-4 md:px-8 py-14 md:py-20 relative z-10'>
          <div className='tech-card rounded-[28px] p-6 md:p-10'>
            <div className='flex flex-col gap-8'>
              <div className='flex flex-wrap items-center gap-3'>
                <Tag color='blue' size='large'>
                  Token Link
                </Tag>
                <Tag color='cyan' size='large'>
                  大模型中转站 · 专业稳定 · 高并发
                </Tag>
              </div>

              <div className='max-w-4xl'>
                <h1 className='text-3xl md:text-5xl font-bold leading-tight text-[#123a72]'>
                  Token Link 专业大模型中转站
                  <br />
                  <span className='text-[#1677ff]'>为生产业务提供稳定与效率</span>
                </h1>
                <p className='mt-5 text-base md:text-lg text-[#355989] leading-8'>
                  统一协议、多供应商可切换、稳定调度与成本优化并行，
                  帮你把精力放在业务创新，而不是对接细节。
                </p>
              </div>

              <div className='grid grid-cols-1 xl:grid-cols-2 gap-5'>
                <div className='rounded-2xl border border-[#9ac8ff] bg-white/85 p-5'>
                  <Text className='!text-[#4a6793] !font-medium'>统一接口地址</Text>
                  <Input
                    readonly
                    value={serverAddress}
                    className='mt-3'
                    size={isMobile ? 'default' : 'large'}
                    suffix={
                      <div className='flex items-center gap-2'>
                        <ScrollList
                          bodyHeight={32}
                          style={{ border: 'unset', boxShadow: 'unset' }}
                        >
                          <ScrollItem
                            mode='wheel'
                            cycled={true}
                            list={endpointItems}
                            selectedIndex={endpointIndex}
                            onSelect={({ index }) => setEndpointIndex(index)}
                          />
                        </ScrollList>
                        <Button
                          type='primary'
                          onClick={handleCopyBaseURL}
                          icon={<IconCopy />}
                          className='!rounded-full'
                        />
                      </div>
                    }
                  />
                </div>

                <div className='rounded-2xl border border-[#9ac8ff] bg-white/85 p-5 flex flex-wrap items-center gap-4'>
                  <Link to='/console'>
                    <Button
                      theme='solid'
                      type='primary'
                      size={isMobile ? 'default' : 'large'}
                      className='!rounded-full !px-8'
                      icon={<IconPlay />}
                    >
                      立即获取密钥
                    </Button>
                  </Link>
                  <Link to='/console/topup'>
                    <Button size={isMobile ? 'default' : 'large'} className='!rounded-full !px-8'>
                      查看套餐与充值
                    </Button>
                  </Link>
                  {docsLink && (
                    <Button
                      size={isMobile ? 'default' : 'large'}
                      className='!rounded-full !px-8'
                      icon={<IconFile />}
                      onClick={() => window.open(docsLink, '_blank')}
                    >
                      文档中心
                    </Button>
                  )}
                </div>
              </div>

              <div className='rounded-2xl border border-[#9ac8ff] bg-white/85 p-5'>
                <div className='flex items-center justify-between flex-wrap gap-3'>
                  <Text className='!text-[#4a6793] !font-medium'>大模型连通率检测</Text>
                  <Button
                    type='primary'
                    theme='solid'
                    icon={<IconRefresh />}
                    loading={checkingConnectivity}
                    onClick={checkConnectivity}
                    className='!rounded-full !px-6'
                  >
                    {checkingConnectivity ? '检测中...' : '检查大模型连通率'}
                  </Button>
                </div>
                <div className='mt-3 text-sm text-[#3e5f8c]'>{connectivitySummary}</div>
                <div className='mt-2'>
                  连通率：
                  <span className='ml-2 text-xl font-bold text-[#1771e6]'>
                    {connectivityRate === null ? '--' : `${connectivityRate}%`}
                  </span>
                </div>
                <div className='mt-3 space-y-2'>
                  {connectivityDetails.map((item) => (
                    <div
                      key={item.key}
                      className='text-sm flex items-center justify-between gap-3 px-3 py-2 rounded-lg bg-[#f2f8ff]'
                    >
                      <span className='text-[#355783]'>{item.name}</span>
                      <span
                        className={`font-medium ${
                          item.ok === null
                            ? 'text-[#6b89b5]'
                            : item.ok
                              ? 'text-[#11a379]'
                              : 'text-[#da5656]'
                        }`}
                      >
                        {item.ok === null
                          ? item.reason
                          : `${item.reason}${item.status ? ` · ${item.status}` : ''}${
                              item.latency ? ` · ${item.latency}ms` : ''
                            }`}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>

          <div className='mt-8 grid grid-cols-1 md:grid-cols-3 gap-4'>
            {coreAdvantages.map((item) => (
              <div key={item.title} className='tech-card rounded-2xl p-5'>
                <h3 className='text-lg font-bold text-[#17406f]'>{item.title}</h3>
                <p className='mt-3 text-sm leading-7 text-[#42628c]'>{item.detail}</p>
              </div>
            ))}
          </div>

          <div className='mt-10'>
            <div className='flex items-center justify-between flex-wrap gap-3 mb-4'>
              <h2 className='text-2xl md:text-3xl font-bold text-[#15437f]'>
                限时特惠分组 · 订阅套餐
              </h2>
              <Tag color='red' size='large'>
                限时特惠
              </Tag>
            </div>

            <div className='grid grid-cols-1 lg:grid-cols-2 gap-6'>
              {packagePlans.map((plan) => (
                <div key={plan.id} className='tech-card rounded-[24px] p-6 md:p-7 relative overflow-hidden'>
                  <div className='absolute top-0 right-0 w-40 h-40 bg-[#7cc4ff]/25 rounded-full blur-2xl' />
                  <div className='relative z-10'>
                    <Tag color='blue' className='!mb-3'>
                      {plan.group}
                    </Tag>
                    <h3 className='text-2xl font-bold text-[#143f77]'>{plan.name}</h3>
                    <p className='mt-3 text-sm text-[#4f6b90] leading-7'>{plan.desc}</p>
                    <div className='mt-5 flex items-end gap-4'>
                      <span className='text-3xl font-bold text-[#1369e8]'>{plan.price}</span>
                      <span className='text-2xl font-extrabold text-[#0f5ac8] token-glow'>
                        {plan.token}
                      </span>
                    </div>
                    <div className='mt-5 space-y-2'>
                      {plan.perks.map((perk) => (
                        <div key={perk} className='text-sm text-[#355783]'>
                          • {perk}
                        </div>
                      ))}
                    </div>
                    <div className='mt-6'>
                      <Link to='/console/topup?view=subscription'>
                        <Button type='primary' theme='solid' className='!rounded-full !px-7'>
                          立即开通该套餐
                        </Button>
                      </Link>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className='mt-10 tech-card rounded-[24px] p-6 md:p-8'>
            <h3 className='text-xl md:text-2xl font-bold text-[#15437f] mb-5'>
              多模型生态与专业调度能力
            </h3>
            <div className='flex flex-wrap items-center gap-6 md:gap-8'>
              {providerLogos.map(({ name, Comp }) => (
                <div key={name} className='flex items-center gap-2'>
                  <Comp size={32} />
                  <span className='text-sm text-[#365982]'>{name}</span>
                </div>
              ))}
              <div className='text-lg font-bold text-[#1b4f8c]'>30+ 模型能力接入</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Home;
