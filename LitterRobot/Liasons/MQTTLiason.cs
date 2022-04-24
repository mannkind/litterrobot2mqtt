using System;
using System.Collections.Generic;
using System.Linq;
using System.Reflection;
using LitterRobot.Models.Options;
using LitterRobot.Models.Shared;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using TwoMQTT;
using TwoMQTT.Interfaces;
using TwoMQTT.Liasons;
using TwoMQTT.Models;
using TwoMQTT.Utils;

namespace LitterRobot.Liasons;

/// <summary>
/// An class representing a managed way to interact with mqtt.
/// </summary>
public class MQTTLiason : MQTTLiasonBase<Resource, Command, SlugMapping, SharedOpts>, IMQTTLiason<Resource, Command>
{
    /// <summary>
    /// Initializes a new instance of the MQTTLiason class.
    /// </summary>
    /// <param name="logger"></param>
    /// <param name="generator"></param>
    /// <param name="sharedOpts"></param>
    public MQTTLiason(ILogger<MQTTLiason> logger, IMQTTGenerator generator, IOptions<SharedOpts> sharedOpts) :
        base(logger, generator, sharedOpts)
    {
        this.DataReceivedExpiration = sharedOpts.Value.DataReceivedExpiration;
    }

    /// <inheritdoc />
    public IEnumerable<(string topic, string payload)> MapData(Resource input)
    {
        var results = new List<(string, string)>();
        var slug = this.Questions
            .Where(x => x.LRID == input.LitterRobotId)
            .Select(x => x.Slug)
            .FirstOrDefault() ?? string.Empty;

        if (string.IsNullOrEmpty(slug))
        {
            this.Logger.LogDebug("Unable to find slug for {litterRobotId}", input.LitterRobotId);
            return results;
        }

        this.Logger.LogDebug("Found slug {slug} for incoming data for {litterRobotId}", slug, input.LitterRobotId);
        results.AddRange(new[]
            {
                this.Generator.DataReceivedTopicPayload(slug),
                
                (this.Generator.StateTopic(slug, nameof(Resource.LitterRobotId)), input.LitterRobotId),
                (this.Generator.StateTopic(slug, nameof(Resource.PowerStatus)), input.PowerStatus),
                (this.Generator.StateTopic(slug, nameof(Resource.UnitStatus)), input.UnitStatus),
                (this.Generator.StateTopic(slug, nameof(Resource.UnitStatusText)), input.UnitStatusText),
                (this.Generator.StateTopic(slug, nameof(Resource.Power)), this.Generator.BooleanOnOff(input.Power)),
                (this.Generator.StateTopic(slug, nameof(Resource.Cycle)), this.Generator.BooleanOnOff(input.Cycle)),
                (this.Generator.StateTopic(slug, nameof(Resource.NightLightActive)), this.Generator.BooleanOnOff(input.NightLightActive)),
                (this.Generator.StateTopic(slug, nameof(Resource.PanelLockActive)), this.Generator.BooleanOnOff(input.PanelLockActive)),
                (this.Generator.StateTopic(slug, nameof(Resource.DFITriggered)), this.Generator.BooleanOnOff(input.DFITriggered)),
                (this.Generator.StateTopic(slug, nameof(Resource.SleepModeActive)), this.Generator.BooleanOnOff(input.SleepModeActive)),
                (this.Generator.StateTopic(slug, nameof(Resource.SleepMode)), input.SleepMode),
            }
        );

        return results;
    }

    /// <inheritdoc />
    public IEnumerable<Command> MapCommand(string topic, string payload)
    {
        var results = new List<Command>();

        foreach (var input in this.Questions)
        {
            var litterRobotId = this.Questions
                .Where(x => x.Slug == input.Slug)
                .Select(x => x.LRID)
                .FirstOrDefault() ?? string.Empty;

            if (string.IsNullOrEmpty(litterRobotId))
            {
                this.Logger.LogDebug("Unable to find litterRobotId for {slug}", input.Slug);
                continue;
            }
            this.Logger.LogDebug("Found {litterRobotId} for incoming data for {slug}", litterRobotId, input.Slug);

            var cmd = new Command
            {
                Command = (int)CommandType.None,
                Data = new Resource
                {
                    LitterRobotId = litterRobotId,
                }
            };

            switch (topic)
            {
                case string s when s == this.Generator.CommandTopic(input.Slug, nameof(Resource.Power)):
                    cmd = cmd with
                    {
                        Command = (int)CommandType.Power,
                        Data = cmd.Data with
                        {
                            Power = payload == Const.ON,
                        }
                    };
                    break;
                case string s when s == this.Generator.CommandTopic(input.Slug, nameof(Resource.Cycle)):
                    cmd = cmd with
                    {
                        Command = (int)CommandType.Cycle,
                        Data = cmd.Data with
                        {
                            Cycle = payload == Const.ON,
                        }
                    };
                    break;
                case string s when s == this.Generator.CommandTopic(input.Slug, nameof(Resource.NightLightActive)):
                    cmd = cmd with
                    {
                        Command = (int)CommandType.NightLight,
                        Data = cmd.Data with
                        {
                            NightLightActive = payload == Const.ON,
                        }
                    };
                    break;
                case string s when s == this.Generator.CommandTopic(input.Slug, nameof(Resource.PanelLockActive)):
                    cmd = cmd with
                    {
                        Command = (int)CommandType.PanelLock,
                        Data = cmd.Data with
                        {
                            PanelLockActive = payload == Const.ON,
                        }
                    };
                    break;
                case string s when s == this.Generator.CommandTopic(input.Slug, nameof(Resource.CleanCycleWaitTimeMinutes)):
                    cmd = cmd with
                    {
                        Command = (int)CommandType.WaitTime,
                        Data = cmd.Data with
                        {
                            CleanCycleWaitTimeMinutes = long.TryParse(payload, out var ccwtm) ? ccwtm : 0,
                        }
                    };
                    break;
                case string s when s == this.Generator.CommandTopic(input.Slug, nameof(Resource.SleepModeActive)):
                    cmd = cmd with
                    {
                        Command = (int)CommandType.Sleep,
                        Data = cmd.Data with
                        {
                            SleepMode = payload,
                        }
                    };
                    break;
            }

            // Skip if an unknown command
            if (cmd.Command == (int)CommandType.None)
            {
                continue;
            }

            results.Add(cmd);
        }

        return results;
    }

    /// <inheritdoc />
    public IEnumerable<string> Subscriptions()
    {
        var topics = new List<string>();
        foreach (var input in this.Questions)
        {
            topics.Add(this.Generator.CommandTopic(input.Slug, nameof(Resource.Power)));
            topics.Add(this.Generator.CommandTopic(input.Slug, nameof(Resource.Cycle)));
            topics.Add(this.Generator.CommandTopic(input.Slug, nameof(Resource.NightLightActive)));
            topics.Add(this.Generator.CommandTopic(input.Slug, nameof(Resource.PanelLockActive)));
            topics.Add(this.Generator.CommandTopic(input.Slug, nameof(Resource.CleanCycleWaitTimeMinutes)));
            topics.Add(this.Generator.CommandTopic(input.Slug, nameof(Resource.SleepModeActive)));
        }

        return topics;
    }

    /// <inheritdoc />
    public IEnumerable<(string slug, string sensor, string type, MQTTDiscovery discovery)> Discoveries()
    {
        var discoveries = new List<(string, string, string, MQTTDiscovery)>();
        var assembly = Assembly.GetAssembly(typeof(Program))?.GetName() ?? new AssemblyName();
        var mapping = new[]
        {
            new { Sensor = nameof(Resource.LitterRobotId), Type = Const.SENSOR, Icon = "mdi:identifier" },
            new { Sensor = nameof(Resource.PowerStatus), Type = Const.SENSOR, Icon = "mdi:power-settings" },
            new { Sensor = nameof(Resource.UnitStatus), Type = Const.SENSOR, Icon = "mdi:robot" },
            new { Sensor = nameof(Resource.UnitStatusText), Type = Const.SENSOR, Icon = "mdi:robot" },
            new { Sensor = nameof(Resource.Power), Type = Const.SWITCH, Icon = "mdi:power" },
            new { Sensor = nameof(Resource.Cycle), Type = Const.SWITCH, Icon = "mdi:rotate-left" },
            new { Sensor = nameof(Resource.NightLightActive), Type = Const.SWITCH, Icon = "mdi:lightbulb" },
            new { Sensor = nameof(Resource.PanelLockActive), Type = Const.SWITCH, Icon = "mdi:lock" },
            new { Sensor = nameof(Resource.DFITriggered), Type = Const.BINARY_SENSOR, Icon = "problem" },
            new { Sensor = nameof(Resource.SleepModeActive), Type = Const.SWITCH, Icon = "mdi:sleep" },
            new { Sensor = nameof(Resource.SleepMode), Type = Const.SENSOR, Icon = "mdi:sleep" },
        };

        foreach (var input in this.Questions)
        {
            discoveries.Add(this.Generator.DataReceivedDiscovery(input.Slug, assembly, this.DataReceivedExpiration));

            foreach (var map in mapping)
            {
                this.Logger.LogDebug("Generating discovery for {litterRobotId} - {sensor}", input.LRID, map.Sensor);
                var discovery = this.Generator.BuildDiscovery(input.Slug, map.Sensor, assembly, false);
                if (map.Type == Const.SWITCH)
                {
                    discovery = discovery with
                    {
                        CommandTopic = this.Generator.CommandTopic(input.Slug, map.Sensor),
                    };
                }

                if (map.Type != Const.BINARY_SENSOR && !string.IsNullOrEmpty(map.Icon))
                {
                    discovery = discovery with
                    {
                        Icon = map.Icon,
                    };
                }
                else if (map.Type == Const.BINARY_SENSOR && !string.IsNullOrEmpty(map.Icon))
                {
                    discovery = discovery with
                    {
                        DeviceClass = map.Icon,
                    };
                }

                discoveries.Add((input.Slug, map.Sensor, map.Type, discovery));
            }
        }

        return discoveries;
    }

    /// <summary>
    /// 
    /// </summary>
    private readonly TimeSpan DataReceivedExpiration;
}
