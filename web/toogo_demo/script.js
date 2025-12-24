const moodMessages = {
    'normal': "你好！我是 Toogo。",
    'happy': "见到你真高兴！汪！",
    'thinking': "正在分析数据...",
    'confused': "嗯？我不明白...",
    'tired': "电量不足...想睡觉...",
    'excited': "太棒了！让我们开始吧！",
    'focused': "目标锁定，正在处理。",
    'sad': "呜...我做错了吗？",
    'conservative': "🛡️ 启动防御协议。本金安全第一。",
    'balanced': "⚖️ 动态平衡中。寻找风险与收益的最佳支点。",
    'aggressive': "🚀 全速推进！目标：超额收益！"
};

let typeWriterInterval;

function typeWriter(element, text, speed = 50) {
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

function setMood(mood) {
    const scene = document.querySelector('.scene');
    const bubble = document.getElementById('chat-bubble');
    
    // 移除所有以 'mood-' 开头的类
    scene.classList.forEach(className => {
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

    console.log(`Mood set to: ${mood}`);
}

// 眼睛跟随鼠标逻辑
function initEyeTracking() {
    const scene = document.querySelector('.scene');
    const eyes = document.querySelectorAll('.eye-ball'); // 获取眼球
    
    if (!scene || eyes.length === 0) return;

    scene.addEventListener('mousemove', (e) => {
        // 如果是闭眼或特殊状态，不跟随
        if (scene.classList.contains('mood-tired') || 
            scene.classList.contains('mood-happy') || 
            scene.classList.contains('mood-sad')) return;

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

        eyes.forEach(eye => {
            eye.style.transform = `translate(${offsetX}px, ${offsetY}px)`;
        });
    });

    // 鼠标离开时复位
    scene.addEventListener('mouseleave', () => {
        eyes.forEach(eye => {
            eye.style.transform = `translate(0, 0)`;
        });
    });
}

// 点击交互逻辑
function initInteractions() {
    const robot = document.getElementById('toogo-robot');
    const head = document.querySelector('.head-group');
    const body = document.querySelector('.body-group');

    if (!robot) return;

    // 点击头部 -> 害羞/开心
    head.addEventListener('click', (e) => {
        e.stopPropagation(); // 防止冒泡
        setMood('happy');
        setTimeout(() => setMood('normal'), 2000);
    });

    // 点击身体 -> 疑惑/震动
    body.addEventListener('click', (e) => {
        e.stopPropagation();
        setMood('excited');
        setTimeout(() => setMood('normal'), 2000);
    });
}

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    console.log("Toogo.Ai Robot Initialized");
    initEyeTracking();
    initInteractions();
});
