import { Group } from "@/types/groupData"
import { useEffect, useState } from "react"
import { API_BASE_URL } from "@/lib/api";


export default function Groups({id}:{id:number}){
    const [group,setGroup]=useState<Group[]>([])
    const getGroup = async()=>{
        try{

            const response= await fetch(`${API_BASE_URL}/api/group/${id}`,{
                credentials:"include"
            }) 
            const data : Group[] = await response.json()
            setGroup(data)
            // if (!response.ok){
                
            // }
        }catch(err){

        }
    }
    useEffect(()=>{
        getGroup()
    },[id])
    return(
        <>
        
        </>
    )

}