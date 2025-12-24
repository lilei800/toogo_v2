<template>
  <div class="toogo-container">
    <header>
      <h1>Toogo.ai</h1>
      <p>您的智能Ai助手 - 土狗</p>
    </header>

    <div class="scene" ref="sceneRef">
      <!-- 聊天气泡 -->
      <div class="chat-bubble" id="chat-bubble" ref="chatBubbleRef">你好！我是 Toogo。</div>

      <!-- 机器人主体 SVG -->
      <svg id="toogo-robot" viewBox="0 0 400 400" xmlns="http://www.w3.org/2000/svg" ref="robotRef">
        <defs>
          <!-- 金属外壳渐变 -->
          <linearGradient id="metal-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" style="stop-color: #ffffff; stop-opacity: 1" />
            <stop offset="50%" style="stop-color: #e0e0e0; stop-opacity: 1" />
            <stop offset="100%" style="stop-color: #b0b0b0; stop-opacity: 1" />
          </linearGradient>
          <!-- 屏幕深色渐变 -->
          <linearGradient id="screen-gradient" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" style="stop-color: #333; stop-opacity: 1" />
            <stop offset="100%" style="stop-color: #111; stop-opacity: 1" />
          </linearGradient>
          <!-- 眼睛发光渐变 -->
          <radialGradient id="eye-glow" cx="50%" cy="50%" r="50%" fx="50%" fy="50%">
            <stop offset="0%" style="stop-color: #00f3ff; stop-opacity: 1" />
            <stop offset="80%" style="stop-color: #00a8ff; stop-opacity: 1" />
            <stop offset="100%" style="stop-color: #0077ff; stop-opacity: 1" />
          </radialGradient>
          <!-- 影子渐变 -->
          <radialGradient id="shadow-gradient" cx="50%" cy="50%" r="50%">
            <stop offset="0%" style="stop-color: #000000; stop-opacity: 0.6" />
            <stop offset="100%" style="stop-color: #000000; stop-opacity: 0" />
          </radialGradient>
          <!-- 护盾渐变 -->
           <radialGradient id="shield-gradient" cx="50%" cy="50%" r="50%">
                <stop offset="0%" style="stop-color:#4cd137;stop-opacity:0" />
                <stop offset="100%" style="stop-color:#4cd137;stop-opacity:0.2" />
            </radialGradient>
        </defs>

        <!-- 阴影 -->
        <ellipse cx="200" cy="360" rx="80" ry="15" class="shadow" fill="url(#shadow-gradient)" />

        <!-- 尾巴 -->
        <g class="tail-group">
          <path
            d="M280 280 Q 320 250 330 200"
            class="tail"
            stroke="#ccc"
            stroke-width="8"
            fill="none"
            stroke-linecap="round"
          />
          <circle cx="330" cy="200" r="8" class="tail-tip" fill="#00f3ff" />
        </g>

        <!-- 身体 -->
        <g class="body-group">
          <rect
            x="140"
            y="220"
            width="120"
            height="100"
            rx="20"
            class="body-chassis"
            fill="url(#metal-gradient)"
          />
          <rect
            x="160"
            y="240"
            width="80"
            height="60"
            rx="10"
            class="body-screen"
            fill="url(#screen-gradient)"
          />
          <text
            x="200"
            y="275"
            text-anchor="middle"
            fill="#00f3ff"
            font-family="Arial"
            font-size="14"
            class="logo-text"
            style="filter: drop-shadow(0 0 2px #00f3ff)"
            >Toogo.Ai</text
          >
          <!-- 数据流效果 (专注模式用) -->
          <g class="data-stream" style="opacity: 0">
            <rect
              x="130"
              y="210"
              width="140"
              height="120"
              rx="25"
              fill="none"
              stroke="#00f3ff"
              stroke-width="2"
              stroke-dasharray="10,5"
            />
          </g>
        </g>

        <!-- 腿 -->
        <g class="legs">
          <rect x="150" y="310" width="20" height="40" rx="10" fill="#bbb" />
          <rect x="230" y="310" width="20" height="40" rx="10" fill="#bbb" />
        </g>

        <!-- 头部 -->
        <g class="head-group">
          <!-- 耳朵 -->
          <g class="ears">
            <path d="M130 150 L 110 100 L 150 120 Z" class="ear left-ear" fill="#ccc" />
            <path d="M270 150 L 290 100 L 250 120 Z" class="ear right-ear" fill="#ccc" />
          </g>

          <!-- 头盔 -->
          <rect
            x="120"
            y="120"
            width="160"
            height="110"
            rx="30"
            class="head-shell"
            fill="url(#metal-gradient)"
          />

          <!-- 脸部屏幕 -->
          <rect x="135" y="135" width="130" height="80" rx="20" class="face-screen" fill="#111" />

          <!-- 眉毛 -->
          <g class="eyebrows">
            <path
              d="M160 160 Q 170 160 180 160"
              class="eyebrow left-eyebrow"
              stroke="#00f3ff"
              stroke-width="3"
              fill="none"
              stroke-linecap="round"
            />
            <path
              d="M220 160 Q 230 160 240 160"
              class="eyebrow right-eyebrow"
              stroke="#00f3ff"
              stroke-width="3"
              fill="none"
              stroke-linecap="round"
            />
          </g>

          <!-- 眼睛 -->
          <g class="eyes">
            <g class="eye-container left-eye" transform="translate(170, 175)">
              <g class="eye">
                <circle
                  cx="0"
                  cy="0"
                  r="12"
                  fill="url(#eye-glow)"
                  class="eye-ball"
                  style="filter: drop-shadow(0 0 5px #00f3ff)"
                />
                <!-- 爱心形状 (默认隐藏) -->
                <path
                  d="M0 5 q-5 -10 -10 -5 t 10 10 t 10 -10 t -10 5"
                  fill="#ff0055"
                  class="eye-heart"
                  style="display: none; transform: scale(1.5) translate(-5px, -5px); filter: drop-shadow(0 0 5px #ff0055)"
                />
                <!-- 眯眼线 (默认隐藏) -->
                <path
                  d="M-10 0 L 10 0"
                  class="eye-squint"
                  stroke="#00f3ff"
                  stroke-width="3"
                  stroke-linecap="round"
                  style="display: none; filter: drop-shadow(0 0 2px #00f3ff)"
                />
              </g>
            </g>
            <g class="eye-container right-eye" transform="translate(230, 175)">
              <g class="eye">
                <circle
                  cx="0"
                  cy="0"
                  r="12"
                  fill="url(#eye-glow)"
                  class="eye-ball"
                  style="filter: drop-shadow(0 0 5px #00f3ff)"
                />
                <path
                  d="M0 5 q-5 -10 -10 -5 t 10 10 t 10 -10 t -10 5"
                  fill="#ff0055"
                  class="eye-heart"
                  style="display: none; transform: scale(1.5) translate(-5px, -5px); filter: drop-shadow(0 0 5px #ff0055)"
                />
                <path
                  d="M-10 0 L 10 0"
                  class="eye-squint"
                  stroke="#00f3ff"
                  stroke-width="3"
                  stroke-linecap="round"
                  style="display: none; filter: drop-shadow(0 0 2px #00f3ff)"
                />
              </g>
            </g>
          </g>

          <!-- 脸颊 (默认隐藏) -->
          <g class="cheeks" style="opacity: 0">
            <ellipse cx="155" cy="190" rx="8" ry="4" fill="#ff0055" opacity="0.6" />
            <ellipse cx="245" cy="190" rx="8" ry="4" fill="#ff0055" opacity="0.6" />
          </g>

          <!-- 嘴巴 (多种形态) -->
          <g class="mouth-group">
            <!-- 微笑 -->
            <path
              d="M185 200 Q 200 210 215 200"
              class="mouth mouth-smile"
              stroke="#00f3ff"
              stroke-width="3"
              fill="none"
              stroke-linecap="round"
              style="opacity: 0"
            />
            <!-- 悲伤/困惑 -->
            <path
              d="M185 205 Q 200 195 215 205"
              class="mouth mouth-sad"
              stroke="#00f3ff"
              stroke-width="3"
              fill="none"
              stroke-linecap="round"
              style="opacity: 0"
            />
            <!-- 张嘴/惊讶 -->
            <ellipse
              cx="200"
              cy="200"
              rx="8"
              ry="8"
              class="mouth mouth-open"
              stroke="#00f3ff"
              stroke-width="3"
              fill="none"
              style="opacity: 0"
            />
            <!-- 直线/严肃 -->
            <path
              d="M190 200 L 210 200"
              class="mouth mouth-line"
              stroke="#00f3ff"
              stroke-width="3"
              fill="none"
              stroke-linecap="round"
              style="opacity: 0"
            />
          </g>

          <!-- 疑惑气泡 (默认隐藏) -->
          <g class="question-mark" style="opacity: 0" transform="translate(280, 100)">
            <text x="0" y="0" fill="#00f3ff" font-size="40" font-weight="bold">?</text>
          </g>

          <!-- 思考指示灯 -->
          <circle cx="260" cy="130" r="4" class="indicator-light" fill="#555" />
        </g>

        <!-- 交易状态特效组件 -->

        <!-- 1. 保守型：能量护盾 -->
        <g class="shield-group" style="opacity: 0; pointer-events: none">
          <path
            d="M110 100 Q 200 40 290 100 V 280 Q 200 380 110 280 Z"
            fill="url(#shield-gradient)"
            stroke="#4cd137"
            stroke-width="2"
            stroke-dasharray="10,5"
            opacity="0.4"
          />
        </g>

        <!-- 2. 激进型：火箭推进火焰 -->
        <g class="fire-group" style="opacity: 0">
          <path
            d="M160 350 Q 170 380 160 400 Q 150 380 160 350"
            fill="#e84118"
            class="flame flame-left"
          />
          <path
            d="M240 350 Q 250 380 240 400 Q 230 380 240 350"
            fill="#e84118"
            class="flame flame-right"
          />
        </g>

        <!-- 3. 平衡型/激进型：全息图表 -->
        <g class="chart-group" style="opacity: 0" transform="translate(280, 180)">
          <polyline
            points="0,50 20,40 40,60 60,20"
            fill="none"
            stroke="#fbc531"
            stroke-width="3"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="chart-line"
          />
          <circle cx="60" cy="20" r="3" fill="#fbc531" class="chart-dot" />
        </g>
      </svg>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { onMounted, onBeforeUnmount, ref, nextTick } from 'vue';

  const sceneRef = ref<HTMLElement | null>(null);
  const chatBubbleRef = ref<HTMLElement | null>(null);
  const robotRef = ref<HTMLElement | null>(null);

  const moodMessages: Record<string, string> = {
    normal: '你好！我是 Toogo。',
    happy: '见到你真高兴！汪！',
    thinking: '正在分析数据...',
    confused: '嗯？我不明白...',
    tired: '电量不足...想睡觉...',
    excited: '太棒了！让我们开始吧！',
    focused: '目标锁定，正在处理。',
    sad: '呜...我做错了吗？',
    conservative: '🛡️ 启动防御协议。本金安全第一。',
    balanced: '⚖️ 动态平衡中。寻找风险与收益的最佳支点。',
    aggressive: '🚀 全速推进！目标：超额收益！',
  };

  let typeWriterInterval: any;
  let moodInterval: any;

  function typeWriter(element: HTMLElement, text: string, speed = 50) {
    if (typeWriterInterval) clearInterval(typeWriterInterval);

    element.textContent = '';
    let i = 0;

    typeWriterInterval = setInterval(() => {
      if (i < text.length) {
        element.textContent += text.charAt(i);
        i++;
      } else {
        clearInterval(typeWriterInterval);
      }
    }, speed);
  }

  function setMood(mood: string) {
    const scene = sceneRef.value;
    const bubble = chatBubbleRef.value;

    if (!scene) return;

    // 移除所有以 'mood-' 开头的类
    scene.classList.forEach((className) => {
      if (className.startsWith('mood-')) {
        scene.classList.remove(className);
      }
    });

    // 如果不是默认状态，添加对应的心情类
    if (mood !== 'normal') {
      scene.classList.add(`mood-${mood}`);
    }

    // 更新气泡文字 (打字机效果)
    if (bubble && moodMessages[mood]) {
      bubble.style.opacity = '1';
      bubble.style.transform = 'translateY(0)';
      typeWriter(bubble, moodMessages[mood]);
    }

    // console.log(`Mood set to: ${mood}`);
  }

  function startAutoMood() {
    const moods = Object.keys(moodMessages);
    let currentIndex = 0;

    // 初始设置
    setMood(moods[currentIndex]);

    moodInterval = setInterval(() => {
        currentIndex = (currentIndex + 1) % moods.length;
        setMood(moods[currentIndex]);
    }, 5000); // 每5秒切换一次
  }

  // 眼睛跟随鼠标逻辑
  function initEyeTracking() {
    const scene = sceneRef.value;
    // const eyes = document.querySelectorAll('.eye-ball'); // Use scoped selector
    if (!scene) return;
    const eyes = scene.querySelectorAll('.eye-ball') as NodeListOf<HTMLElement>;

    if (eyes.length === 0) return;

    scene.addEventListener('mousemove', (e) => {
      // 如果是闭眼或特殊状态，不跟随
      if (
        scene.classList.contains('mood-tired') ||
        scene.classList.contains('mood-happy') ||
        scene.classList.contains('mood-sad')
      )
        return;

      const rect = scene.getBoundingClientRect();
      const centerX = rect.left + rect.width / 2;
      const centerY = rect.top + rect.height / 2;

      const mouseX = e.clientX;
      const mouseY = e.clientY;

      // 计算角度和距离
      const angle = Math.atan2(mouseY - centerY, mouseX - centerX);
      const distance = Math.min(10, Math.hypot(mouseX - centerX, mouseY - centerY) / 10); // 限制最大移动距离

      const offsetX = Math.cos(angle) * distance;
      const offsetY = Math.sin(angle) * distance;

      eyes.forEach((eye) => {
        eye.style.transform = `translate(${offsetX}px, ${offsetY}px)`;
      });
    });

    // 鼠标离开时复位
    scene.addEventListener('mouseleave', () => {
      eyes.forEach((eye) => {
        eye.style.transform = `translate(0, 0)`;
      });
    });
  }

  // 点击交互逻辑
  function initInteractions() {
    const robot = robotRef.value;
    if (!robot) return;
    
    const head = robot.querySelector('.head-group') as HTMLElement;
    const body = robot.querySelector('.body-group') as HTMLElement;


    // 点击头部 -> 害羞/开心
    if (head) {
        head.addEventListener('click', (e) => {
        e.stopPropagation(); // 防止冒泡
        setMood('happy');
        setTimeout(() => setMood('normal'), 2000);
        });
    }

    // 点击身体 -> 疑惑/震动
    if (body) {
        body.addEventListener('click', (e) => {
        e.stopPropagation();
        setMood('excited');
        setTimeout(() => setMood('normal'), 2000);
        });
    }
  }

  onMounted(() => {
    // console.log('Toogo.Ai Robot Initialized');
    initEyeTracking();
    initInteractions();
    startAutoMood();
  });

  onBeforeUnmount(() => {
    if (typeWriterInterval) clearInterval(typeWriterInterval);
    if (moodInterval) clearInterval(moodInterval);
  });
</script>

<style lang="less" scoped>
  :root {
    --primary-color: #00f3ff;
    --secondary-color: #e0e0e0;
    --bg-color: #1a1a2e;
    --panel-bg: #16213e;
    --text-color: #fff;
    --heart-color: #ff0055;
    --sad-color: #4a69bd;
    --excited-color: #f6b93b;
  }
  
  /* Helper vars since we can't use :root in scoped style effectively for vars unless global.
     We will replace vars with values or define host vars.
  */
  .toogo-container {
    --primary-color: #2080f0;
    --secondary-color: #666;
    --bg-color: transparent;
    --panel-bg: rgba(255, 255, 255, 0.8);
    --text-color: #333;
    --heart-color: #ff0055;
    --sad-color: #4a69bd;
    --excited-color: #f6b93b;

    /* background-color: var(--bg-color); */ /* Don't set bg on container to blend with login page */
    color: var(--text-color);
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    height: 100%;
    width: 100%;
  }


  header h1 {
    margin: 0;
    font-size: 2.5rem;
    text-transform: uppercase;
    letter-spacing: 2px;
    color: var(--primary-color);
    text-shadow: 0 0 10px rgba(32, 128, 240, 0.3);
    text-align: center;
  }

  header p {
      text-align: center;
  }

  .scene {
    position: relative;
    height: 400px;
    width: 100%; /* adjust width */
    max-width: 600px;
    display: flex;
    justify-content: center;
    align-items: center;
    margin: 20px 0;
    background: transparent;
    /* border-radius: 20px; */
    /* overflow: hidden; */
    perspective: 1000px; /* 为3D效果做准备 */
  }

  /* --- 动态科技背景 --- */
  /* .tech-background { ... } removed */

  /* --- 聊天气泡 --- */
  .chat-bubble {
    position: absolute;
    top: 40px;
    right: 40px;
    background: rgba(255, 255, 255, 0.8);
    backdrop-filter: blur(5px);
    border: 1px solid var(--primary-color);
    color: var(--primary-color);
    padding: 10px 20px;
    border-radius: 15px 15px 0 15px;
    font-size: 14px;
    z-index: 10;
    opacity: 0; /* 初始隐藏，由JS控制显示 */
    transform: translateY(0);
    box-shadow: 0 0 10px rgba(32, 128, 240, 0.2);
    max-width: 150px;
    min-height: 20px; /* 防止文字为空时塌陷 */
    display: flex;
    align-items: center;
  }

  /* 打字机光标效果 */
  .chat-bubble::after {
    content: '|';
    animation: blink-cursor 1s infinite;
    margin-left: 2px;
    color: var(--primary-color);
  }

  @keyframes blink-cursor {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0;
    }
  }

  #toogo-robot {
    width: 300px;
    height: 300px;
    transition: all 0.5s ease;
    z-index: 1; /* 确保在背景之上 */
    cursor: pointer; /* 提示可点击 */
  }

  .shadow {
    transition: all 0.5s ease;
    transform-origin: center;
    opacity: 0.8;
  }

  /* --- 基础部件样式 --- */
  .head-group {
    transform-origin: 200px 200px;
    transition: transform 0.5s ease;
  }

  .eyebrow {
    transition: all 0.3s ease;
    transform-origin: center;
  }

  .mouth {
    transition: opacity 0.3s ease;
  }

  .cheeks {
    transition: opacity 0.3s ease;
  }

  .tail-group {
    transform-origin: 280px 280px;
    transition: transform 0.5s ease;
  }

  .eye-ball {
    transition: all 0.3s ease;
  }

  .indicator-light {
    transition: fill 0.3s;
  }

  /* --- 默认动画 (Idle) --- */
  @keyframes float {
    0%,
    100% {
      transform: translateY(0);
    }
    50% {
      transform: translateY(-5px);
    }
  }

  @keyframes blink {
    0%,
    96%,
    98% {
      transform: scaleY(1);
    }
    97% {
      transform: scaleY(0.1);
    }
  }

  #toogo-robot {
    animation: float 3s ease-in-out infinite;
  }

  .eyes {
    transform-origin: center;
  }

  .eye-container {
    transform-origin: center;
  }

  .eye {
    animation: blink 4s infinite;
    transform-origin: 0 0; /* 修改为 0 0，因为内容中心就在原点 */
  }

  /* --- 心情状态: 开心 (Happy) --- */
  /* Need deep selectors because SVG is complex? No, just normal scoped styles if elements are inside. */
  /* Note: In scoped CSS, we need to match elements inside child components or use :deep(). But here everything is in the same component. */
  
  :deep(.mood-happy) #toogo-robot, .mood-happy #toogo-robot {
    animation: jump 0.8s infinite alternate;
  }

  :deep(.mood-happy) .tail-group, .mood-happy .tail-group {
    animation: wag 0.5s infinite ease-in-out;
  }

  :deep(.mood-happy) .eye-ball, .mood-happy .eye-ball {
    display: none;
  }

  :deep(.mood-happy) .eye-squint, .mood-happy .eye-squint {
    display: block !important;
    transform: scale(1.2);
  }

  :deep(.mood-happy) .mouth-smile, .mood-happy .mouth-smile {
    opacity: 1 !important;
  }

  :deep(.mood-happy) .cheeks, .mood-happy .cheeks {
    opacity: 1 !important;
  }

  :deep(.mood-happy) .eyebrow, .mood-happy .eyebrow {
    transform: translateY(-5px);
  }

  :deep(.mood-happy) .shadow, .mood-happy .shadow {
    animation: shadow-jump 0.8s infinite alternate;
  }

  @keyframes jump {
    0% {
      transform: translateY(0);
    }
    100% {
      transform: translateY(-20px);
    }
  }

  @keyframes shadow-jump {
    0% {
      transform: scale(1);
      opacity: 0.8;
    }
    100% {
      transform: scale(0.6);
      opacity: 0.4;
    }
  }

  @keyframes wag {
    0% {
      transform: rotate(0deg);
    }
    25% {
      transform: rotate(15deg);
    }
    75% {
      transform: rotate(-15deg);
    }
    100% {
      transform: rotate(0deg);
    }
  }

  /* --- 心情状态: 思考 (Thinking) --- */
  :deep(.mood-thinking) .head-group, .mood-thinking .head-group {
    transform: rotate(15deg);
  }

  :deep(.mood-thinking) .indicator-light, .mood-thinking .indicator-light {
    animation: flash-light 0.5s infinite;
    fill: #ff9f43;
  }

  :deep(.mood-thinking) .eye-ball, .mood-thinking .eye-ball {
    transform: scale(1.2);
  }

  :deep(.mood-thinking) .mouth-line, .mood-thinking .mouth-line {
    opacity: 1 !important;
  }

  :deep(.mood-thinking) .left-eyebrow, .mood-thinking .left-eyebrow {
    transform: translateY(-3px) rotate(-10deg);
  }

  :deep(.mood-thinking) .right-eyebrow, .mood-thinking .right-eyebrow {
    transform: translateY(3px) rotate(10deg);
  }

  @keyframes flash-light {
    0%,
    100% {
      fill: #555;
    }
    50% {
      fill: #ff9f43;
      box-shadow: 0 0 10px #ff9f43;
    }
  }

  /* --- 心情状态: 疑惑 (Confused) --- */
  :deep(.mood-confused) .head-group, .mood-confused .head-group {
    animation: shake-head 2s infinite ease-in-out;
  }

  :deep(.mood-confused) .question-mark, .mood-confused .question-mark {
    opacity: 1 !important;
    animation: float-up 2s infinite;
  }

  :deep(.mood-confused) .right-ear, .mood-confused .right-ear {
    transform: rotate(20deg) translate(10px, -10px);
  }

  :deep(.mood-confused) .mouth-open, .mood-confused .mouth-open {
    opacity: 1 !important;
    transform: scale(0.8);
  }

  :deep(.mood-confused) .left-eyebrow, .mood-confused .left-eyebrow {
    transform: translateY(-5px);
  }

  :deep(.mood-confused) .right-eyebrow, .mood-confused .right-eyebrow {
    transform: translateY(2px);
  }

  @keyframes shake-head {
    0%,
    100% {
      transform: rotate(0deg);
    }
    25% {
      transform: rotate(-10deg);
    }
    75% {
      transform: rotate(10deg);
    }
  }

  @keyframes float-up {
    0% {
      transform: translate(280px, 100px);
      opacity: 0;
    }
    50% {
      opacity: 1;
    }
    100% {
      transform: translate(280px, 80px);
      opacity: 0;
    }
  }

  /* --- 心情状态: 疲惫 (Tired) --- */
  :deep(.mood-tired) #toogo-robot, .mood-tired #toogo-robot {
    animation: float 6s ease-in-out infinite; /* Slower float */
    filter: grayscale(0.5);
  }

  :deep(.mood-tired) .eye, .mood-tired .eye {
    animation: sleepy-blink 3s infinite;
  }

  :deep(.mood-tired) .head-group, .mood-tired .head-group {
    transform: translateY(10px) rotate(-5deg);
  }

  :deep(.mood-tired) .mouth-open, .mood-tired .mouth-open {
    opacity: 1 !important;
    animation: yawn 3s infinite;
  }

  :deep(.mood-tired) .eyebrow, .mood-tired .eyebrow {
    transform: rotate(10deg) translateY(2px); /* 八字眉 */
  }

  @keyframes sleepy-blink {
    0%,
    50%,
    100% {
      transform: scaleY(0.3);
    } /* Half closed */
    25% {
      transform: scaleY(0.1);
    } /* Closed */
    75% {
      transform: scaleY(0.3);
    }
  }

  @keyframes yawn {
    0%,
    100% {
      transform: scale(0.5);
    }
    50% {
      transform: scale(1.2);
    }
  }

  /* --- 心情状态: 兴奋 (Excited) --- */
  :deep(.mood-excited) #toogo-robot, .mood-excited #toogo-robot {
    animation: vibrate 0.2s infinite;
  }

  :deep(.mood-excited) .face-screen, .mood-excited .face-screen {
    animation: rainbow-bg 1s infinite;
  }

  :deep(.mood-excited) .tail-group, .mood-excited .tail-group {
    animation: wag 0.2s infinite;
  }

  :deep(.mood-excited) .eye-ball, .mood-excited .eye-ball {
    display: none;
  }

  :deep(.mood-excited) .eye-heart, .mood-excited .eye-heart {
    display: block !important;
    animation: pulse-heart 0.4s infinite alternate;
  }

  :deep(.mood-excited) .mouth-smile, .mood-excited .mouth-smile {
    opacity: 1 !important;
    transform: scale(1.2);
  }

  :deep(.mood-excited) .cheeks, .mood-excited .cheeks {
    opacity: 1 !important;
  }

  @keyframes vibrate {
    0% {
      transform: translate(0, 0);
    }
    25% {
      transform: translate(2px, 2px);
    }
    50% {
      transform: translate(-2px, -2px);
    }
    75% {
      transform: translate(2px, -2px);
    }
    100% {
      transform: translate(-2px, 2px);
    }
  }

  @keyframes rainbow-bg {
    0% {
      fill: #111;
    }
    25% {
      fill: #222;
    }
    50% {
      fill: #333;
    }
    75% {
      fill: #222;
    }
    100% {
      fill: #111;
    }
  }

  @keyframes pulse-heart {
    0% {
      transform: scale(1.5);
    }
    100% {
      transform: scale(1.8);
    }
  }

  /* --- 心情状态: 专注 (Focused) --- */
  :deep(.mood-focused) .data-stream, .mood-focused .data-stream {
    opacity: 1 !important;
    animation: spin-data 4s linear infinite;
  }

  :deep(.mood-focused) .eye-ball, .mood-focused .eye-ball {
    fill: #ff3333; /* Red focus */
    r: 8; /* Smaller pupil */
  }

  :deep(.mood-focused) .head-group, .mood-focused .head-group {
    transform: translateZ(0); /* Fix for sharpness if needed */
  }

  :deep(.mood-focused) .mouth-line, .mood-focused .mouth-line {
    opacity: 1 !important;
  }

  :deep(.mood-focused) .eyebrow, .mood-focused .eyebrow {
    transform: rotate(15deg) translateY(5px); /* 紧皱 */
  }

  @keyframes spin-data {
    0% {
      stroke-dashoffset: 0;
    }
    100% {
      stroke-dashoffset: 100;
    }
  }

  /* --- 心情状态: 失落 (Sad) --- */
  :deep(.mood-sad) .head-group, .mood-sad .head-group {
    transform: translateY(20px) rotate(-10deg);
  }

  :deep(.mood-sad) .tail-group, .mood-sad .tail-group {
    transform: rotate(40deg); /* Tail down */
  }

  :deep(.mood-sad) .face-screen, .mood-sad .face-screen {
    fill: #0a0a15;
  }

  :deep(.mood-sad) .eye-ball, .mood-sad .eye-ball {
    fill: var(--sad-color);
  }

  :deep(.mood-sad) #toogo-robot, .mood-sad #toogo-robot {
    transform: scale(0.95);
  }

  :deep(.mood-sad) .mouth-sad, .mood-sad .mouth-sad {
    opacity: 1 !important;
  }

  :deep(.mood-sad) .eyebrow, .mood-sad .eyebrow {
    transform: rotate(-10deg) translateY(-2px); /* 悲伤眉 */
  }

  /* --- 交易状态: 保守型 (Conservative) --- */
  :deep(.mood-conservative) .shield-group, .mood-conservative .shield-group {
    opacity: 1 !important;
    animation: shield-pulse 3s infinite ease-in-out;
  }

  :deep(.mood-conservative) .eye-ball, .mood-conservative .eye-ball {
    fill: #4cd137; /* 绿色安全 */
  }

  :deep(.mood-conservative) .mouth-line, .mood-conservative .mouth-line {
    opacity: 1 !important;
    stroke: #4cd137;
  }

  :deep(.mood-conservative) .body-screen, .mood-conservative .body-screen {
    fill: #1e272e; /* 深沉背景 */
  }

  :deep(.mood-conservative) .eyebrow, .mood-conservative .eyebrow {
    stroke: #4cd137;
  }

  @keyframes shield-pulse {
    0%,
    100% {
      opacity: 0.4;
      transform: scale(1);
    }
    50% {
      opacity: 0.7;
      transform: scale(1.02);
    }
  }

  /* --- 交易状态: 平衡型 (Balanced) --- */
  :deep(.mood-balanced) .head-group, .mood-balanced .head-group {
    animation: balance-head 4s infinite ease-in-out;
  }

  :deep(.mood-balanced) .eye-ball, .mood-balanced .eye-ball {
    fill: #00a8ff; /* 蓝色冷静 */
  }

  :deep(.mood-balanced) .chart-group, .mood-balanced .chart-group {
    opacity: 1 !important;
  }

  :deep(.mood-balanced) .chart-line, .mood-balanced .chart-line {
    stroke: #00a8ff;
    animation: draw-chart 2s infinite alternate;
  }

  :deep(.mood-balanced) .chart-dot, .mood-balanced .chart-dot {
    fill: #00a8ff;
  }

  :deep(.mood-balanced) .mouth-smile, .mood-balanced .mouth-smile {
    opacity: 1 !important;
    stroke: #00a8ff;
    transform: scale(0.8);
  }

  @keyframes balance-head {
    0%,
    100% {
      transform: rotate(0deg);
    }
    25% {
      transform: rotate(5deg);
    }
    75% {
      transform: rotate(-5deg);
    }
  }

  @keyframes draw-chart {
    0% {
      stroke-dasharray: 0, 100;
    }
    100% {
      stroke-dasharray: 100, 0;
    }
  }

  /* --- 交易状态: 激进型 (Aggressive) --- */
  :deep(.mood-aggressive) #toogo-robot, .mood-aggressive #toogo-robot {
    animation: aggressive-shake 0.1s infinite;
    transform: translateY(-20px); /* 升空 */
  }

  :deep(.mood-aggressive) .fire-group, .mood-aggressive .fire-group {
    opacity: 1 !important;
  }

  :deep(.mood-aggressive) .flame, .mood-aggressive .flame {
    animation: fire-burn 0.1s infinite alternate;
  }

  :deep(.mood-aggressive) .eye-ball, .mood-aggressive .eye-ball {
    fill: #e84118; /* 红色激进 */
    filter: drop-shadow(0 0 8px #e84118);
  }

  :deep(.mood-aggressive) .eyebrow, .mood-aggressive .eyebrow {
    stroke: #e84118;
    transform: rotate(20deg) translateY(10px); /* 怒目/专注 */
  }

  :deep(.mood-aggressive) .mouth-smile, .mood-aggressive .mouth-smile {
    opacity: 1 !important;
    stroke: #e84118;
    transform: scale(1.2); /* 自信大笑 */
  }

  :deep(.mood-aggressive) .chart-group, .mood-aggressive .chart-group {
    opacity: 1 !important;
    transform: translate(280px, 150px) scale(1.5); /* 更大的图表 */
  }

  :deep(.mood-aggressive) .chart-line, .mood-aggressive .chart-line {
    stroke: #e84118;
  }

  :deep(.mood-aggressive) .chart-dot, .mood-aggressive .chart-dot {
    fill: #e84118;
  }

  :deep(.mood-aggressive) .shadow, .mood-aggressive .shadow {
    transform: scale(0.6);
    opacity: 0.4;
    animation: shadow-shake 0.1s infinite;
  }

  @keyframes aggressive-shake {
    0% {
      transform: translate(0, -20px);
    }
    25% {
      transform: translate(1px, -19px);
    }
    50% {
      transform: translate(-1px, -21px);
    }
    75% {
      transform: translate(1px, -21px);
    }
    100% {
      transform: translate(-1px, -19px);
    }
  }

  @keyframes shadow-shake {
    0%,
    100% {
      transform: scale(0.6) translate(0, 0);
    }
    50% {
      transform: scale(0.6) translate(1px, 0);
    }
  }

  @keyframes fire-burn {
    0% {
      transform: scaleY(1);
      opacity: 0.8;
    }
    100% {
      transform: scaleY(1.2);
      opacity: 1;
    }
  }

  /* --- 控制面板 --- */
  /* .controls removed */

  /* button removed */

  /* button:hover removed */

  /* @media removed */
</style>

