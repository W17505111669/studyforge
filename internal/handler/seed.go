package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// SeedData 一键生成示例学习材料和分析结果（用于演示和测试）
func (h *Handler) SeedData(c *gin.Context) {
	userID := c.GetString("userID")

	// 检查是否已有 seed 数据（通过标题匹配）
	var count int64
	h.DB.Model(&model.Material{}).Where("user_id = ? AND title = ?", userID, "【示例】Go 并发编程入门").Count(&count)
	if count > 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "示例数据已存在，无需重复生成",
			"seeded":  false,
		})
		return
	}

	now := time.Now()

	// ========== 材料 1：Go 并发编程 ==========
	mat1 := model.Material{
		UserID:      userID,
		Title:       "【示例】Go 并发编程入门",
		ContentType: "text",
		Content: `Go 语言天生支持并发编程，其核心机制是 goroutine 和 channel。

## Goroutine
Goroutine 是 Go 语言中的轻量级线程。使用 go 关键字即可启动一个新的 goroutine：

go func() {
    fmt.Println("Hello from goroutine!")
}()

Goroutine 的创建成本极低（初始栈仅 2KB），一个 Go 程序可以轻松创建数十万个 goroutine。

## Channel
Channel 是 goroutine 之间通信的管道。声明和初始化：

ch := make(chan int)    // 无缓冲 channel
ch := make(chan int, 5) // 缓冲 channel（容量 5）

发送和接收：
ch <- 42    // 发送
val := <-ch // 接收

## Select 多路复用
select 语句用于在多个 channel 操作中选择就绪的一个：

select {
case msg := <-ch1:
    fmt.Println(msg)
case ch2 <- 42:
    fmt.Println("sent")
case <-time.After(time.Second):
    fmt.Println("timeout")
}

## sync 包
sync 包提供了基本的同步原语：
- sync.Mutex: 互斥锁
- sync.WaitGroup: 等待一组 goroutine 完成
- sync.Once: 确保函数只执行一次
- sync.Map: 并发安全的 map

## 并发模式
常见并发模式包括：
1. Worker Pool（工作池）：固定数量的 goroutine 处理任务队列
2. Fan-in/Fan-out（扇入扇出）：多个 goroutine 读写同一 channel
3. Pipeline（管道）：将处理阶段串联，每个阶段是一个 goroutine`,
		Status: "completed",
		AnalyzedAt: &now,
	}

	mat1Analysis := map[string]interface{}{
		"summary": "本文全面介绍了 Go 语言的并发编程模型，从基础的 goroutine 和 channel 到高级的 select 多路复用和 sync 包，以及常见的并发设计模式。",
		"key_points": []map[string]string{
			{"concept": "Goroutine", "detail": "Go 的轻量级线程，初始栈仅 2KB，用 go 关键字启动", "difficulty": "easy"},
			{"concept": "Channel", "detail": "goroutine 间通信管道，分有缓冲和无缓冲两种", "difficulty": "medium"},
			{"concept": "Select", "detail": "多路复用语句，在多个 channel 操作中选择就绪的一个", "difficulty": "medium"},
			{"concept": "sync 包", "detail": "提供 Mutex、WaitGroup、Once、Map 等同步原语", "difficulty": "medium"},
			{"concept": "并发模式", "detail": "Worker Pool、Fan-in/Fan-out、Pipeline 等常见设计模式", "difficulty": "hard"},
		},
		"relationships": []map[string]string{
			{"from": "Goroutine", "to": "Channel", "type": "related"},
			{"from": "Channel", "to": "Select", "type": "prerequisite"},
			{"from": "sync 包", "to": "Goroutine", "type": "related"},
			{"from": "并发模式", "to": "Channel", "type": "prerequisite"},
		},
		"importance": "Go 并发编程是 Go 语言的核心优势之一，掌握 goroutine 和 channel 是编写高效 Go 程序的基础。",
	}
	analysisJSON1, _ := json.Marshal(mat1Analysis)
	mat1.AnalysisData = string(analysisJSON1)

	mat1Graph := map[string]interface{}{
		"nodes": []map[string]interface{}{
			{"name": "Goroutine", "category": 0, "symbolSize": 50},
			{"name": "Channel", "category": 0, "symbolSize": 50},
			{"name": "Select", "category": 1, "symbolSize": 35},
			{"name": "sync.Mutex", "category": 1, "symbolSize": 30},
			{"name": "sync.WaitGroup", "category": 1, "symbolSize": 30},
			{"name": "Worker Pool", "category": 2, "symbolSize": 40},
			{"name": "Fan-in/Fan-out", "category": 2, "symbolSize": 35},
			{"name": "Pipeline", "category": 2, "symbolSize": 35},
		},
		"edges": []map[string]string{
			{"source": "Goroutine", "target": "Channel", "label": "通信"},
			{"source": "Channel", "target": "Select", "label": "多路复用"},
			{"source": "Goroutine", "target": "sync.Mutex", "label": "同步"},
			{"source": "Goroutine", "target": "sync.WaitGroup", "label": "等待"},
			{"source": "Channel", "target": "Worker Pool", "label": "任务分发"},
			{"source": "Channel", "target": "Fan-in/Fan-out", "label": "扇入扇出"},
			{"source": "Channel", "target": "Pipeline", "label": "管道串联"},
		},
	}
	graphJSON1, _ := json.Marshal(mat1Graph)
	mat1.GraphData = string(graphJSON1)

	// ========== 材料 2：机器学习基础 ==========
	mat2 := model.Material{
		UserID:      userID,
		Title:       "【示例】机器学习基础概念",
		ContentType: "text",
		Content: `机器学习是人工智能的核心分支，让计算机能够从数据中自动学习和改进，而无需显式编程。

## 监督学习
监督学习使用标注数据进行训练。模型学习输入到输出的映射关系。

### 分类（Classification）
预测离散类别标签。例如：垃圾邮件检测、图像识别。
常见算法：逻辑回归、决策树、SVM、随机森林、神经网络。

### 回归（Regression）
预测连续数值。例如：房价预测、温度预测。
常见算法：线性回归、岭回归、多项式回归。

## 无监督学习
无监督学习使用未标注的数据，发现数据中的隐藏结构。

### 聚类（Clustering）
将数据分成若干组。例如：客户分群、图像分割。
常见算法：K-Means、DBSCAN、层次聚类。

### 降维（Dimensionality Reduction）
减少特征数量，保留重要信息。例如：PCA、t-SNE。

## 模型评估
关键评估指标：
- 准确率（Accuracy）：正确预测的比例
- 精确率（Precision）：预测为正的样本中真正为正的比例
- 召回率（Recall）：真正为正的样本中被正确预测的比例
- F1 Score：精确率和召回率的调和平均

## 过拟合与欠拟合
过拟合：模型在训练数据上表现好，但在测试数据上表现差（记忆了噪声）。
欠拟合：模型在训练数据上就表现不好（没有学到规律）。

解决方法：
- 过拟合：增加数据、正则化、Dropout、早停
- 欠拟合：增加特征、减小正则化、使用更复杂模型

## 梯度下降
梯度下降是最常用的优化算法，通过计算损失函数关于参数的梯度来更新模型参数。

学习率 α 控制每次更新的步长：
θ = θ - α × ∇L(θ)

变体包括：SGD（随机梯度下降）、Mini-batch GD、Adam、RMSprop。`,
		Status: "completed",
		AnalyzedAt: &now,
	}

	mat2Analysis := map[string]interface{}{
		"summary": "本文系统介绍了机器学习的基础概念，涵盖监督学习（分类/回归）、无监督学习（聚类/降维）、模型评估指标、过拟合与欠拟合的对策，以及梯度下降优化算法。",
		"key_points": []map[string]string{
			{"concept": "监督学习", "detail": "使用标注数据训练模型，包括分类和回归两大任务", "difficulty": "easy"},
			{"concept": "无监督学习", "detail": "使用未标注数据发现隐藏结构，包括聚类和降维", "difficulty": "easy"},
			{"concept": "模型评估", "detail": "准确率、精确率、召回率、F1 Score 等评估指标", "difficulty": "medium"},
			{"concept": "过拟合与欠拟合", "detail": "模型泛化能力的两个极端，需要平衡", "difficulty": "medium"},
			{"concept": "梯度下降", "detail": "通过计算梯度更新参数的优化算法，学习率控制步长", "difficulty": "hard"},
		},
		"relationships": []map[string]string{
			{"from": "监督学习", "to": "模型评估", "type": "related"},
			{"from": "无监督学习", "to": "监督学习", "type": "related"},
			{"from": "过拟合与欠拟合", "to": "模型评估", "type": "prerequisite"},
			{"from": "梯度下降", "to": "监督学习", "type": "prerequisite"},
		},
		"importance": "机器学习基础概念是深入学习 AI 领域的必备知识，理解这些核心概念有助于选择合适的算法和调优模型。",
	}
	analysisJSON2, _ := json.Marshal(mat2Analysis)
	mat2.AnalysisData = string(analysisJSON2)

	mat2Graph := map[string]interface{}{
		"nodes": []map[string]interface{}{
			{"name": "监督学习", "category": 0, "symbolSize": 50},
			{"name": "分类", "category": 0, "symbolSize": 35},
			{"name": "回归", "category": 0, "symbolSize": 35},
			{"name": "无监督学习", "category": 1, "symbolSize": 45},
			{"name": "聚类", "category": 1, "symbolSize": 30},
			{"name": "降维", "category": 1, "symbolSize": 30},
			{"name": "模型评估", "category": 2, "symbolSize": 40},
			{"name": "过拟合", "category": 2, "symbolSize": 35},
			{"name": "欠拟合", "category": 2, "symbolSize": 35},
			{"name": "梯度下降", "category": 3, "symbolSize": 45},
		},
		"edges": []map[string]string{
			{"source": "监督学习", "target": "分类", "label": "包含"},
			{"source": "监督学习", "target": "回归", "label": "包含"},
			{"source": "无监督学习", "target": "聚类", "label": "包含"},
			{"source": "无监督学习", "target": "降维", "label": "包含"},
			{"source": "监督学习", "target": "模型评估", "label": "评估"},
			{"source": "模型评估", "target": "过拟合", "label": "诊断"},
			{"source": "模型评估", "target": "欠拟合", "label": "诊断"},
			{"source": "梯度下降", "target": "监督学习", "label": "优化"},
		},
	}
	graphJSON2, _ := json.Marshal(mat2Graph)
	mat2.GraphData = string(graphJSON2)

	// ========== 保存材料 ==========
	if err := h.DB.Create(&mat1).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建示例材料 1 失败: " + err.Error()})
		return
	}
	if err := h.DB.Create(&mat2).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建示例材料 2 失败: " + err.Error()})
		return
	}

	// ========== 卡片 ==========
	cards := []model.Card{
		// Go 并发编程卡片
		{UserID: userID, MaterialID: mat1.ID, Concept: "Goroutine", Detail: "Go 语言中的轻量级线程，使用 go 关键字启动。初始栈仅 2KB，可轻松创建数十万个。", Formula: "go func() { ... }()", MemoryTip: "想象成一个超级轻量的\"小工人\"，几乎不占资源就能启动", Difficulty: "easy", Tags: "go,并发,goroutine"},
		{UserID: userID, MaterialID: mat1.ID, Concept: "Channel", Detail: "Goroutine 之间通信的管道。无缓冲 channel 同步通信，有缓冲 channel 异步通信。", Formula: "ch := make(chan int, 5) // 缓冲\nch <- 42 // 发送\nval := <-ch // 接收", MemoryTip: "把 channel 想象成一根\"水管\"，数据从一端流入，另一端流出", Difficulty: "medium", Tags: "go,并发,channel,通信"},
		{UserID: userID, MaterialID: mat1.ID, Concept: "Select 多路复用", Detail: "在多个 channel 操作中选择就绪的一个，类似 switch 但用于 channel。", Formula: "select {\ncase msg := <-ch1: ...\ncase ch2 <- val: ...\ncase <-time.After(d): ...\n}", MemoryTip: "想象站在多个水管的交叉口，哪个来水了就处理哪个", Difficulty: "medium", Tags: "go,并发,select,channel"},
		{UserID: userID, MaterialID: mat1.ID, Concept: "Worker Pool 模式", Detail: "固定数量的 goroutine 从任务 channel 中消费任务，实现并发任务处理。", Formula: "for w := 0; w < numWorkers; w++ {\n  go func() {\n    for task := range jobs {\n      results <- process(task)\n    }\n  }()\n}", MemoryTip: "像一个工厂里固定数量的工人在流水线上工作", Difficulty: "hard", Tags: "go,并发,模式,worker"},
		// 机器学习卡片
		{UserID: userID, MaterialID: mat2.ID, Concept: "监督学习 vs 无监督学习", Detail: "监督学习使用标注数据学习输入到输出的映射；无监督学习从无标注数据中发现隐藏结构。", MemoryTip: "监督学习像有老师批改作业；无监督学习像自己探索规律", Difficulty: "easy", Tags: "ml,监督学习,无监督学习,基础"},
		{UserID: userID, MaterialID: mat2.ID, Concept: "分类与回归", Detail: "分类预测离散类别（如垃圾邮件检测），回归预测连续数值（如房价预测）。", MemoryTip: "分类是\"选 A 还是 B\"，回归是\"猜一个数字\"", Difficulty: "easy", Tags: "ml,分类,回归,监督学习"},
		{UserID: userID, MaterialID: mat2.ID, Concept: "精确率与召回率", Detail: "精确率 = 预测正例中真正例的比例；召回率 = 真正例中被找到的比例。F1 是二者的调和平均。", Formula: "P = TP/(TP+FP)\nR = TP/(TP+FN)\nF1 = 2×P×R/(P+R)", MemoryTip: "精确率：抓对的比例；召回率：该抓的抓到了多少", Difficulty: "medium", Tags: "ml,评估,精确率,召回率"},
		{UserID: userID, MaterialID: mat2.ID, Concept: "过拟合", Detail: "模型在训练集上表现好但测试集上差，说明模型\"记住了\"训练数据的噪声而非学到通用规律。", MemoryTip: "像死记硬背的学生，课本都会但换题就不会了", Difficulty: "medium", Tags: "ml,过拟合,泛化"},
		{UserID: userID, MaterialID: mat2.ID, Concept: "梯度下降", Detail: "通过计算损失函数关于参数的梯度来迭代更新模型参数，学习率 α 控制步长。", Formula: "θ = θ - α × ∇L(θ)", MemoryTip: "想象蒙眼下山，每一步沿着最陡的下坡方向走", Difficulty: "hard", Tags: "ml,优化,梯度下降,学习率"},
	}

	for i := range cards {
		if err := h.DB.Create(&cards[i]).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建示例卡片失败: " + err.Error()})
			return
		}
	}

	// ========== 练习题 ==========
	quizzes := []model.Quiz{
		// Go 并发编程练习题
		{
			UserID: userID, MaterialID: mat1.ID, Type: "choice", Difficulty: "easy",
			Question: "Go 语言中启动 goroutine 的关键字是什么？",
			Options:  `["go", "goroutine", "spawn", "thread"]`,
			Answer:   "go",
			Explanation: "Go 使用 go 关键字后跟函数调用来启动一个新的 goroutine。",
		},
		{
			UserID: userID, MaterialID: mat1.ID, Type: "choice", Difficulty: "medium",
			Question: "以下关于无缓冲 Channel 的描述，正确的是？",
			Options:  `["发送和接收同时阻塞直到配对成功", "发送不会阻塞", "接收不会阻塞", "容量默认为 1"]`,
			Answer:   "发送和接收同时阻塞直到配对成功",
			Explanation: "无缓冲 channel 要求发送方和接收方同时就绪，否则阻塞。这实现了同步通信。",
		},
		{
			UserID: userID, MaterialID: mat1.ID, Type: "fill", Difficulty: "medium",
			Question: "创建一个容量为 5 的整型缓冲 Channel 的代码是：ch := make(chan int, ___)",
			Answer:   "5",
			Explanation: "make(chan int, 5) 创建容量为 5 的缓冲 channel，可容纳 5 个未接收的值。",
		},
		{
			UserID: userID, MaterialID: mat1.ID, Type: "choice", Difficulty: "hard",
			Question: "在 Worker Pool 模式中，通常使用什么来分发任务给多个 worker？",
			Options:  `["Channel", "全局变量", "文件", "HTTP 请求"]`,
			Answer:   "Channel",
			Explanation: "Worker Pool 通常使用 channel 作为任务队列，多个 goroutine worker 从 channel 中消费任务。",
		},
		// 机器学习练习题
		{
			UserID: userID, MaterialID: mat2.ID, Type: "choice", Difficulty: "easy",
			Question: "垃圾邮件检测属于机器学习的哪类任务？",
			Options:  `["分类", "回归", "聚类", "降维"]`,
			Answer:   "分类",
			Explanation: "垃圾邮件检测是二分类任务：将邮件分为\"垃圾\"或\"非垃圾\"两类。",
		},
		{
			UserID: userID, MaterialID: mat2.ID, Type: "choice", Difficulty: "medium",
			Question: "模型在训练集上准确率 99%，但测试集上只有 60%，这最可能是？",
			Options:  `["过拟合", "欠拟合", "数据不足", "学习率太大"]`,
			Answer:   "过拟合",
			Explanation: "训练集和测试集表现差距大是过拟合的典型特征，模型记住了训练数据的噪声。",
		},
		{
			UserID: userID, MaterialID: mat2.ID, Type: "fill", Difficulty: "hard",
			Question: "梯度下降的参数更新公式为 θ = θ - ___ × ∇L(θ)，其中空白处应填什么？",
			Answer:   "α",
			Explanation: "α（alpha）是学习率，控制每次参数更新的步长大小。",
		},
		{
			UserID: userID, MaterialID: mat2.ID, Type: "short_answer", Difficulty: "medium",
			Question: "请简述精确率（Precision）和召回率（Recall）的区别。",
			Answer:   "精确率衡量预测为正的样本中真正为正的比例，关注预测的准确性；召回率衡量所有真正为正的样本中被正确预测出来的比例，关注查找的完整性。",
			Explanation: "精确率 = TP/(TP+FP)，关注\"预测的正例有多少是对的\"；召回率 = TP/(TP+FN)，关注\"所有正例中找到了多少\"。",
		},
	}

	for i := range quizzes {
		if err := h.DB.Create(&quizzes[i]).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建示例练习题失败: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "示例数据生成成功！",
		"seeded":   true,
		"materials": 2,
		"cards":    len(cards),
		"quizzes":  len(quizzes),
	})
}
