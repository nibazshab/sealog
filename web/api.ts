import axios from "axios"

export const base = ""

interface Topic {
    id: number
    created_at: number
    title: string
    mode_id: string
}

interface Post {
    topic_id: number
    floor: number
    updated_at: string
    content: string
}

interface Mode {
    id: number
    name: string
    pub: boolean
}

interface Result<T> {
    code: number
    msg: string
    data?: T
}

const req = axios.create({
    baseURL: base,
    headers: {
        "Content-Type": "application/json;charset=utf-8"
    },
    withCredentials: false
})

req.interceptors.request.use(config => {
    const token = localStorage.getItem("token")
    if (token) {
        config.headers.Authorization = token
    }
    return config
})

req.interceptors.response.use(response => response.data)

export const getAv = (
    id: string
): Promise<Result<{
    topic: Topic
    posts: Post[]
}>> => {
    return req.get("/av/" + id)
}

export const getCvs = (): Promise<Result<Mode[]>> => {
    return req.get("/cv/")
}

export const getAvs = (
    offset?: number
): Promise<Result<Topic[]>> => {
    return req.get("/av/", {
        params: offset != null ? {offset} : undefined
    })
}

export const getAvsByCv = (
    id: string,
    offset?: number
): Promise<Result<Topic[]>> => {
    return req.get("/cv/" + id, {
        params: offset != null ? {offset} : undefined
    })
}

export const getUp = (): Promise<Result<number>> => {
    return req.get("/up")
}

export const createCv = (
    name: string,
    deep: number
): Promise<Result<Mode>> => {
    return req.post("/api/cv/create", {
        name,
        deep
    })
}

export const deleteCv = (
    id: number
): Promise<Result<void>> => {
    return req.post("/api/cv/delete", {
        id
    })
}

export const updateCv = (
    id: number,
    name: string,
    deep: number
): Promise<Result<Mode>> => {
    return req.post("/api/cv/update", {
        id,
        name,
        deep
    })
}

export const createAv = (
    title: string,
    mode_id: number,
    content: string
): Promise<Result<Topic>> => {
    return req.post("/api/av/create", {
        title,
        mode_id,
        content
    })
}

export const deleteAv = (
    id: number
): Promise<Result<void>> => {
    return req.post("/api/av/delete", {
        id
    })
}

export const updateAv = (
    id: number,
    title: string,
    mode_id: number
): Promise<Result<Topic>> => {
    return req.post("/api/av/update", {
        id,
        title,
        mode_id
    })
}

export const createFl = (
    topic_id: number,
    content: string
): Promise<Result<Post>> => {
    return req.post("/api/fl/create", {
        topic_id,
        content
    })
}

export const deleteFl = (
    topic_id: number,
    floor: number
): Promise<Result<void>> => {
    return req.post("/api/fl/delete", {
        topic_id,
        floor
    })
}

export const updateFl = (
    topic_id: number,
    floor: number,
    content: string
): Promise<Result<Post>> => {
    return req.post("/api/fl/update", {
        topic_id,
        floor,
        content
    })
}

export const login = (
    password: string
): Promise<Result<string>> => {
    return req.post("/auth/login", {
        password
    })
}

export const changeUp = (
    password: string
): Promise<Result<void>> => {
    return req.post("/api/up/change", {
        password
    })
}
