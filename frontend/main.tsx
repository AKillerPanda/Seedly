import React, { useState, useRef, useEffect } from 'react'
import ReactDOM from 'react-dom/client'
import './styles.css'

interface Message {
  id: string
  type: 'user' | 'assistant'
  content: string
  timestamp: Date
  action?: string
}

interface PortfolioMetrics {
  totalBalance?: number
  monthlyInvestment?: number
  returnRate?: number
  investedAmount?: number
}

function InvestMateApp() {
  const wsUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws'
  const [messages, setMessages] = useState<Message[]>([
    {
      id: '0',
      type: 'assistant',
      content: "👋 Hi! I'm InvestMate, your friendly AI investment advisor. I'm here to help you understand investing, plan your wealth-building journey, and set up automated investing. What would you like to know?",
      timestamp: new Date(),
    }
  ])
  const [inputValue, setInputValue] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [metrics, setMetrics] = useState<PortfolioMetrics>({})
  const [ws, setWs] = useState<WebSocket | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  // Initialize WebSocket connection
  useEffect(() => {
    const connect = () => {
      try {
        const socket = new WebSocket(wsUrl)
        
        socket.onopen = () => {
          console.log('Connected to InvestMate')
          setWs(socket)
        }
        
        socket.onmessage = (event) => {
          const data = JSON.parse(event.data)
          if (data.type === 'assistant_message') {
            const newMessage: Message = {
              id: Date.now().toString(),
              type: 'assistant',
              content: data.content || 'Unable to process response',
              timestamp: new Date(),
              action: data.requires_confirmation ? 'confirm' : undefined,
            }
            setMessages(prev => [...prev, newMessage])
            setIsLoading(false)
          }
        }
        
        socket.onerror = (error) => {
          console.error('WebSocket error:', error)
          setMessages(prev => [...prev, {
            id: Date.now().toString(),
            type: 'assistant',
            content: '⚠️ Connection error. Please check if the server is running on ws://localhost:8080/ws',
            timestamp: new Date(),
          }])
          setIsLoading(false)
        }
        
        socket.onclose = () => {
          console.log('Disconnected from InvestMate')
          setWs(null)
          setTimeout(connect, 3000) // Reconnect after 3 seconds
        }
      } catch (error) {
        console.error('Connection failed:', error)
      }
    }
    
    connect()
    
    return () => {
      if (ws) ws.close()
    }
  }, [wsUrl])

  // Auto-scroll to bottom
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const sendMessage = (e: React.FormEvent) => {
    e.preventDefault()
    if (!inputValue.trim() || !ws || ws.readyState !== WebSocket.OPEN) return

    const userMessage: Message = {
      id: Date.now().toString(),
      type: 'user',
      content: inputValue,
      timestamp: new Date(),
    }

    setMessages(prev => [...prev, userMessage])
    setInputValue('')
    setIsLoading(true)

    try {
      ws.send(JSON.stringify({
        type: 'message',
        content: inputValue,
      }))
    } catch (error) {
      console.error('Failed to send message:', error)
      setIsLoading(false)
    }

    inputRef.current?.focus()
  }

  const suggestedPrompts = [
    "Help me start investing",
    "What's my investment profile?",
    "Show me investment projections",
    "How can I save more money?"
  ]

  const handleSuggestedPrompt = (prompt: string) => {
    setInputValue(prompt)
  }

  return (
    <div className="investmate-container">
      {/* Sidebar */}
      <aside className="sidebar">
        <div className="sidebar-header">
          <div className="logo">
            <span className="logo-icon">💰</span>
            <span className="logo-text">InvestMate</span>
          </div>
          <p className="tagline">AI Investment Advisor</p>
        </div>

        <div className="metrics-section">
          <h3>Quick Stats</h3>
          <div className="metric-card">
            <div className="metric-label">Portfolio Status</div>
            <div className="metric-value">Ready to Explore</div>
          </div>
          <div className="metric-card">
            <div className="metric-label">Features Available</div>
            <div className="metric-value">21 Tools</div>
          </div>
          <div className="metric-card">
            <div className="metric-label">Your Advisor</div>
            <div className="metric-value">Claude AI</div>
          </div>
        </div>

        <div className="quick-actions">
          <h3>Quick Actions</h3>
          <button className="action-button">📊 View Profile</button>
          <button className="action-button">🎯 Set Goal</button>
          <button className="action-button">📈 See Projections</button>
          <button className="action-button">🤖 Automate Investing</button>
        </div>

        <div className="sidebar-footer">
          <p className="footer-text">Powered by Claude AI & Liminal Banking APIs</p>
        </div>
      </aside>

      {/* Main Chat Area */}
      <main className="chat-container">
        {/* Chat Header */}
        <div className="chat-header">
          <div className="header-content">
            <h1>Welcome to InvestMate</h1>
            <p>Your friendly AI investment advisor</p>
          </div>
          <div className="connection-status">
            <span className={`status-dot ${ws?.readyState === WebSocket.OPEN ? 'connected' : 'disconnected'}`}></span>
            <span className="status-text">{ws?.readyState === WebSocket.OPEN ? 'Connected' : 'Connecting...'}</span>
          </div>
        </div>

        {/* Messages Area */}
        <div className="messages-area">
          {messages.length === 1 && (
            <div className="empty-state">
              <div className="empty-icon">✨</div>
              <h2>Let's Start Your Investing Journey</h2>
              <p>Ask me anything about investing, creating goals, or automating your wealth-building</p>
              
              <div className="suggested-prompts">
                <p className="prompt-label">Try asking:</p>
                {suggestedPrompts.map((prompt) => (
                  <button
                    key={prompt}
                    className="prompt-chip"
                    onClick={() => handleSuggestedPrompt(prompt)}
                  >
                    {prompt}
                  </button>
                ))}
              </div>
            </div>
          )}

          {messages.map((msg) => (
            <div key={msg.id} className={`message-group message-${msg.type}`}>
              <div className={`message-bubble ${msg.type}`}>
                <div className="message-content">
                  {msg.type === 'assistant' ? (
                    <p>{msg.content}</p>
                  ) : (
                    <p>{msg.content}</p>
                  )}
                </div>
                <span className="message-time">
                  {msg.timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </span>
              </div>
            </div>
          ))}

          {isLoading && (
            <div className="message-group message-assistant">
              <div className="message-bubble assistant loading">
                <div className="typing-indicator">
                  <span></span>
                  <span></span>
                  <span></span>
                </div>
              </div>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        {/* Input Area */}
        <form className="input-area" onSubmit={sendMessage}>
          <div className="input-wrapper">
            <input
              ref={inputRef}
              type="text"
              placeholder="Ask me about investing, goals, or your portfolio..."
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              disabled={!ws || ws.readyState !== WebSocket.OPEN || isLoading}
              className="message-input"
              autoFocus
            />
            <button
              type="submit"
              disabled={!ws || ws.readyState !== WebSocket.OPEN || isLoading || !inputValue.trim()}
              className="send-button"
              title="Send message (Enter)"
            >
              <span className="send-icon">→</span>
            </button>
          </div>
          <p className="input-hint">Press Enter to send • Your data is secure and private</p>
        </form>
      </main>
    </div>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(<InvestMateApp />)
