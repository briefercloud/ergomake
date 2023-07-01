import {
  ChevronLeftIcon,
  ExclamationTriangleIcon,
} from '@heroicons/react/24/solid'
import AnsiToHTML from 'ansi-to-html'
import React, { useCallback } from 'react'
import ReactMarkdown from 'react-markdown'
import { Link } from 'react-router-dom'

import Background from '../components/Background'
import ButtonGroup from '../components/ButtonGroup'
import ExternalLinkIcon from '../components/ExternalLinkIcon'
import Navbar from '../components/Navbar'
import Pane from '../components/Pane'
import XCircleIcon from '../components/XCircleIcon'
import { Environment, hasLogs } from '../hooks/useEnvironment'
import { LogData } from '../hooks/useLogs'
import { Owner } from '../hooks/useOwners'
import { Profile } from '../hooks/useProfile'
import { Repo } from '../hooks/useRepo'

const converter = new AnsiToHTML()

interface Props {
  profile: Profile
  owner: Owner
  repo: Repo
  environment: Environment
  logsSwitch: 'live' | 'build'
  onToggleLogsSwitch: () => void
  buildLogs: LogData
  liveLogs: LogData
  currentService: number
  onChangeService: (idx: number) => void
}
function EnvironmentLayout(props: Props) {
  const { logsSwitch, onToggleLogsSwitch } = props
  const onChangeLogsSwitch = useCallback(
    (ls: string) => {
      if (ls !== logsSwitch) {
        onToggleLogsSwitch()
      }
    },
    [logsSwitch, onToggleLogsSwitch]
  )

  const showingServices = props.environment.services.filter((s) =>
    props.logsSwitch === 'build' ? s.build !== '' : true
  )

  const currentService =
    showingServices[Math.min(props.currentService, showingServices.length - 1)]
  const logData = (
    props.logsSwitch === 'live' ? props.liveLogs : props.buildLogs
  )[currentService?.id ?? '']
  const logs = (logData ?? [])
    .slice()
    .reverse()
    .map((l, i) => {
      let html = `\u001b[90m[${new Date(
        l.timestamp
      ).toISOString()}]\u001b[0m\t${l.message}`
      try {
        html = converter.toHtml(html)
      } catch (e) {
        // look away
      }

      return (
        <pre
          key={i}
          className="text-white whitespace-pre-wrap"
          dangerouslySetInnerHTML={{ __html: html }}
        />
      )
    })

  return (
    <Background>
      <Navbar profile={props.profile} currentOwner={props.owner.login} />
      <div className="px-20 pt-8 space-y-8">
        <Link
          className="flex items-center text-primary-500 font-bold text-sm"
          to={`/gh/${props.owner.login}/repos/${props.repo.name}`}
        >
          <ChevronLeftIcon className="w-5 h-5 text-primary-500 mr-1" />
          Back to project
        </Link>
        <Pane>
          <div className="flex items-center justify-between">
            <h2 className="flex items-center text-white font-bold text-sm">
              <img
                className="mr-2 w-8 h-8 rounded-full"
                src={props.owner.avatar}
                alt={props.owner.login}
              />
              {props.owner.login}
              <span className="mx-2 text-primary-500">/</span>
              {props.repo.name}
              <span className="mx-2 text-primary-500">/</span>
              {props.environment.branch}
            </h2>
            {props.environment.status === 'degraded' && (
              <div className="flex items-center">
                <XCircleIcon />
                <p className="ml-2 text-sm font-bold text-red-400">Failed</p>
              </div>
            )}
            {props.environment.status === 'limited' && (
              <div className="flex items-center">
                <XCircleIcon />
                <p className="ml-2 text-sm font-bold text-red-400">Limited</p>
              </div>
            )}
          </div>
          <h1 className="mt-4 text-white font-bold text-4xl">
            {props.environment.branch}
          </h1>

          {(props.environment.status === 'success' ||
            props.environment.status === 'stale') && (
            <div className="mt-8">
              <p>Link to preview</p>
              <div className="flex justify-between border-2 border-outcolor mt-2">
                <p className="mx-4 my-3">
                  https://{props.environment.services[0]?.url ?? ''}
                </p>
                <div>
                  <a
                    href={`https://${props.environment.services[0]?.url ?? ''}`}
                    target="_blank"
                    rel="noreferrer"
                    className="flex items-center h-full border-l-2 border-outcolor px-3"
                  >
                    <ExternalLinkIcon />
                  </a>
                </div>
              </div>
            </div>
          )}
        </Pane>
        {props.environment.degradedReason && (
          <div className="text-gray-300 p-4 rounded-lg flex space-x-2 items-center justify-center">
            <ExclamationTriangleIcon className="w-6 h-6 inline-block text-yellow-500" />
            <ReactMarkdown className="prose prose-invert">
              {props.environment.degradedReason.message}
            </ReactMarkdown>
          </div>
        )}

        {hasLogs(props.environment) && (
          <Pane className="flex mt-8 flex-col">
            <h2 className="text-white font-bold text-lg mb-8">Logs</h2>
            <ButtonGroup
              value={props.logsSwitch}
              onChange={onChangeLogsSwitch}
              className="ml-auto bg-black"
            >
              <ButtonGroup.Item value="live">Live</ButtonGroup.Item>
              <ButtonGroup.Item value="build">Build</ButtonGroup.Item>
            </ButtonGroup>
            <div className="relative">
              {showingServices.map((s, i) => {
                let className = 'min-w-[100px] py-2 px-4 font-bold'
                if (s.id === currentService?.id) {
                  className +=
                    ' relative text-primary-500 border-2 border-b-0 rounded-t-xl z-10 bg-black border-outcolor'
                } else {
                  className += ' text-gray-300'
                }

                return (
                  <button
                    key={s.id}
                    onClick={() => props.onChangeService(i)}
                    className={className}
                    style={{ marginBottom: -2 }}
                  >
                    {s.name}
                  </button>
                )
              })}
              <div
                className={`h-[42rem] bg-black flex flex-col-reverse border-2 rounded-xl py-2 px-4 border-outcolor overflow-y-auto scrollbar-hide${
                  props.currentService === 0 ? ' rounded-tl-none' : ''
                }`}
              >
                {logs}
              </div>
            </div>
          </Pane>
        )}
      </div>
    </Background>
  )
}

export default EnvironmentLayout
