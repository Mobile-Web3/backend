package http

// CreateMnemonic godoc
// @Summary      Создание мнемоника
// @Tags         account
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body account.CreateMnemonicInput true "body"
// @Success      200 {object} apiResponse{result=string}
// @Router       /account/mnemonic [post]
func CreateMnemonic() {}

// CreateAccount godoc
// @Summary      Получение аккаунта по мнемонику
// @Tags         account
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body account.CreateAccountInput true "body"
// @Success      200 {object} apiResponse{result=account.KeyResponse}
// @Router       /account/create [post]
func CreateAccount() {}

// RestoreAccount godoc
// @Summary      Получение аккаунта по ключу
// @Tags         account
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body account.RestoreAccountInput true "body"
// @Success      200 {object} apiResponse{result=account.KeyResponse}
// @Router       /account/restore [post]
func RestoreAccount() {}

// CheckBalance godoc
// @Summary      Получить инфу о балансе
// @Tags         account
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body account.BalanceInput true "body"
// @Success      200 {object} apiResponse{result=account.BalanceResponse}
// @Router       /account/balance [post]
func CheckBalance() {}

// GetAllChains godoc
// @Summary      Получение данных о сетях
// @Tags         chain
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @Success      200 {object} apiResponse{result=[]chain.ShortResponse}
// @Router       /chains/all [post]
func GetAllChains() {}

// SendTransaction godoc
// @Summary      Отправить транзакцию
// @Tags         transaction
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body transaction.SendInput true "body"
// @Success      200 {object} apiResponse{result=transaction.SendResponse}
// @Router       /transaction/send [post]
func SendTransaction() {}

// SendTransactionFirebase godoc
// @Summary      Отправить транзакцию с подпиской на события тендерминта с пушами в firebase
// @Tags         transaction
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body transaction.SendInputFirebase true "body"
// @Success      200 {object} apiResponse{result=transaction.SendResponse}
// @Router       /transaction/send/firebase [post]
func SendTransactionFirebase() {}

// SimulateTransaction godoc
// @Summary      Симуляция транзакции для расчета параметров
// @Tags         transaction
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body transaction.SimulateInput true "body"
// @Success      200 {object} apiResponse{result=transaction.SimulateResponse}
// @Router       /transaction/simulate [post]
func SimulateTransaction() {}
