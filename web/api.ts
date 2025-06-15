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

interface ResAid {
    topic: Topic
    posts: Post[]
}

interface ResCid {
    mode: Mode
    topics: Topic[]
}

const req = axios.create({
    baseURL: base + "/api",
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

export const search = (
    q: string
): Result<Topic> => {
    return req.get("/search", {
        params: {q}
    })
}

export const reqAv = (
    offset?: number
): Promise<Result<Topic[]>> => {
    return req.get("/av", {
        params: offset != null ? {offset} : undefined
    })
}

export const reqCv = (): Promise<Result<Mode[]>> => {
    return req.get("/cv")
}

export const reqAid = (
    aid: string
): Promise<Result<ResAid>> => {
    return req.get("/av/" + aid)
}

export const reqCid = (
    cid: string,
    offset?: number
): Promise<Result<ResCid>> => {
    return req.get("/cv/" + cid, {
        params: offset != null ? {offset} : undefined
    })
}

export const createCv = (
    name: string,
    deep: number
): Promise<Result<Mode>> => {
    return req.post("/cv/create", {
        name,
        deep
    })
}

export const deleteCv = (
    id: number
): Promise<Result<void>> => {
    return req.post("/cv/delete", {
        id
    })
}

export const updateCv = (
    id: number,
    name: string,
    deep: number
): Promise<Result<Mode>> => {
    return req.post("/cv/update", {
        id,
        name,
        deep
    })
}

export const createAv = (
    title: string,
    mode_id: number,
    content: string
): Promise<Result<ResAid>> => {
    return req.post("/av/create", {
        title,
        mode_id,
        content
    })
}

export const deleteAv = (
    id: number
): Promise<Result<void>> => {
    return req.post("/av/delete", {
        id
    })
}

export const updateAv = (
    id: number,
    title: string,
    mode_id: number
): Promise<Result<Topic>> => {
    return req.post("/av/update", {
        id,
        title,
        mode_id
    })
}

export const createFl = (
    topic_id: number,
    content: string
): Promise<Result<Post>> => {
    return req.post("/fl/create", {
        topic_id,
        content
    })
}

export const deleteFl = (
    topic_id: number,
    floor: number
): Promise<Result<void>> => {
    return req.post("/fl/delete", {
        topic_id,
        floor
    })
}

export const updateFl = (
    topic_id: number,
    floor: number,
    content: string
): Promise<Result<Post>> => {
    return req.post("/fl/update", {
        topic_id,
        floor,
        content
    })
}

export const reqSpace = (): Promise<Result<boolean>> => {
    return req.get("/space")
}

export const login = (
    password: string
): Promise<Result<string>> => {
    return req.post("/login", {
        password
    })
}

export const changeAuth = (
    password: string
): Promise<Result<void>> => {
    return req.post("/auth/change", {
        password
    })
}
