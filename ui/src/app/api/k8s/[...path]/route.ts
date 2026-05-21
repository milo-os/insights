import { NextRequest, NextResponse } from "next/server"
import { existsSync, readFileSync } from "fs"
import https from "https"

/**
 * API proxy route that forwards requests to the Kubernetes API.
 *
 * In-cluster: Uses service account token and CA certificate for direct authentication.
 * Local dev: Falls back to kubectl-proxy for easier development.
 */

const SERVICE_ACCOUNT_TOKEN_PATH = "/var/run/secrets/kubernetes.io/serviceaccount/token"
const SERVICE_ACCOUNT_CA_PATH = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
const KUBECTL_PROXY_URL = process.env.KUBECTL_PROXY_URL || "http://127.0.0.1:8001"

interface K8sConfig {
  baseUrl: string
  token?: string
  httpsAgent?: https.Agent
}

/**
 * Get Kubernetes API configuration based on runtime environment.
 * Returns in-cluster config if running in a pod, otherwise kubectl-proxy config for local dev.
 */
function getK8sConfig(): K8sConfig {
  // Check if running in-cluster by looking for service account token
  const isInCluster = existsSync(SERVICE_ACCOUNT_TOKEN_PATH)

  if (isInCluster) {
    // In-cluster: read token and CA cert
    const token = readFileSync(SERVICE_ACCOUNT_TOKEN_PATH, "utf8").trim()
    const ca = readFileSync(SERVICE_ACCOUNT_CA_PATH)

    // Build API server URL from environment variables
    const host = process.env.KUBERNETES_SERVICE_HOST || "kubernetes.default.svc"
    const port = process.env.KUBERNETES_SERVICE_PORT || "443"
    const baseUrl = `https://${host}:${port}`

    // Create HTTPS agent with cluster CA for TLS verification
    const httpsAgent = new https.Agent({
      ca: ca,
      rejectUnauthorized: true,
    })

    return { baseUrl, token, httpsAgent }
  }

  // Local development: use kubectl-proxy (no auth needed)
  return { baseUrl: KUBECTL_PROXY_URL }
}

/**
 * Make a request to the Kubernetes API with appropriate authentication.
 */
async function makeK8sRequest(
  url: string,
  options: RequestInit
): Promise<Response> {
  const config = getK8sConfig()

  // For HTTPS requests with custom agent, use node-fetch or native fetch with agent support
  // Next.js 13+ uses native fetch which doesn't support httpsAgent, so we need to use https module directly
  if (config.httpsAgent) {
    return new Promise((resolve, reject) => {
      const parsedUrl = new URL(url)

      // Build headers object for https request
      const headers: Record<string, string> = {
        ...(options.headers as Record<string, string>),
      }

      // Add Authorization header if we have a token
      if (config.token) {
        headers["Authorization"] = `Bearer ${config.token}`
      }

      const requestOptions: https.RequestOptions = {
        hostname: parsedUrl.hostname,
        port: parsedUrl.port,
        path: parsedUrl.pathname + parsedUrl.search,
        method: options.method || "GET",
        headers: headers,
        agent: config.httpsAgent,
      }

      const req = https.request(requestOptions, (res) => {
        let data = ""

        res.on("data", (chunk) => {
          data += chunk
        })

        res.on("end", () => {
          // Construct a Response-like object
          const response = new Response(data, {
            status: res.statusCode,
            statusText: res.statusMessage,
            headers: res.headers as HeadersInit,
          })
          resolve(response)
        })
      })

      req.on("error", reject)

      if (options.body) {
        req.write(options.body)
      }

      req.end()
    })
  }

  // Local development: use standard fetch
  // Add Authorization header if we have a token
  if (config.token) {
    const headers = new Headers(options.headers)
    headers.set("Authorization", `Bearer ${config.token}`)
    options.headers = headers
  }

  return fetch(url, options)
}

export async function GET(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  const config = getK8sConfig()
  const path = params.path.join("/")
  const url = `${config.baseUrl}/${path}${request.nextUrl.search}`

  try {
    const response = await makeK8sRequest(url, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    })

    const data = await response.json()
    return NextResponse.json(data, { status: response.status })
  } catch (error) {
    console.error("API proxy error:", error)
    return NextResponse.json(
      { message: "Failed to connect to Kubernetes API" },
      { status: 502 }
    )
  }
}

export async function POST(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  const config = getK8sConfig()
  const path = params.path.join("/")
  const url = `${config.baseUrl}/${path}`
  const body = await request.json()

  try {
    const response = await makeK8sRequest(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    })

    const data = await response.json()
    return NextResponse.json(data, { status: response.status })
  } catch (error) {
    console.error("API proxy error:", error)
    return NextResponse.json(
      { message: "Failed to connect to Kubernetes API" },
      { status: 502 }
    )
  }
}

export async function PUT(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  const config = getK8sConfig()
  const path = params.path.join("/")
  const url = `${config.baseUrl}/${path}`
  const body = await request.json()

  try {
    const response = await makeK8sRequest(url, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    })

    const data = await response.json()
    return NextResponse.json(data, { status: response.status })
  } catch (error) {
    console.error("API proxy error:", error)
    return NextResponse.json(
      { message: "Failed to connect to Kubernetes API" },
      { status: 502 }
    )
  }
}

export async function PATCH(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  const config = getK8sConfig()
  const path = params.path.join("/")
  const url = `${config.baseUrl}/${path}`
  const body = await request.json()

  try {
    const response = await makeK8sRequest(url, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/merge-patch+json",
      },
      body: JSON.stringify(body),
    })

    const data = await response.json()
    return NextResponse.json(data, { status: response.status })
  } catch (error) {
    console.error("API proxy error:", error)
    return NextResponse.json(
      { message: "Failed to connect to Kubernetes API" },
      { status: 502 }
    )
  }
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  const config = getK8sConfig()
  const path = params.path.join("/")
  const url = `${config.baseUrl}/${path}`

  try {
    const response = await makeK8sRequest(url, {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
      },
    })

    if (response.status === 204) {
      return new NextResponse(null, { status: 204 })
    }

    const data = await response.json()
    return NextResponse.json(data, { status: response.status })
  } catch (error) {
    console.error("API proxy error:", error)
    return NextResponse.json(
      { message: "Failed to connect to Kubernetes API" },
      { status: 502 }
    )
  }
}
