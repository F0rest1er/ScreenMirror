document.addEventListener('DOMContentLoaded', () => {
    const fullScreenBtn = document.getElementById('fullScreenBtn');
    const bodyEl = document.body;
    
    fullScreenBtn.addEventListener('click', () => {
        if (!document.fullscreenElement) {
            document.documentElement.requestFullscreen().then(() => {
                bodyEl.classList.add('is-fullscreen');
            }).catch(err => {
                console.error(`Ошибка при переходе в полноэкранный режим: ${err.message} (${err.name})`);
            });
        } else {
            document.exitFullscreen().then(() => {
                bodyEl.classList.remove('is-fullscreen');
            });
        }
    });

    document.addEventListener('fullscreenchange', () => {
        if (!document.fullscreenElement) {
            bodyEl.classList.remove('is-fullscreen');
        }
    });

    const videoNode = document.querySelector('.viewer-content__video');
    const loaderNode = document.querySelector('.viewer-content__loader');

    if (videoNode && loaderNode) {
        videoNode.addEventListener('load', () => {
            loaderNode.style.display = 'none';
        });
        videoNode.addEventListener('error', () => {
            loaderNode.textContent = 'Ошибка загрузки. Проверьте права "Запись экрана" на Mac.';
            loaderNode.style.display = 'block';
        });
    }
});
