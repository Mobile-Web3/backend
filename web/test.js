function initUserId() {
    if (window.location.href.includes('user_id')) {
        return
    }

    const userIdDiv = document.getElementById('hidden-user-id')
    const querySymbol = window.location.href.includes('?') ? '&' : '?'
    window.history.pushState(history.state, '', `${querySymbol}user_id=${userIdDiv.innerText}`)
}

initUserId()
