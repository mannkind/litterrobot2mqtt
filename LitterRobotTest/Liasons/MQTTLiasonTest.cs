using System.Linq;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using Microsoft.VisualStudio.TestTools.UnitTesting;
using Moq;
using TwoMQTT.Core.Utils;
using LitterRobot.Liasons;
using LitterRobot.Models.Options;
using LitterRobot.Models.Shared;
using System;

namespace LitterRobotTest.Liasons
{
    [TestClass]
    public class MQTTLiasonTest
    {
        [TestMethod]
        public void MapDataTest()
        {
            var tests = new[] {
                new {
                    Q = new SlugMapping { LRID = BasicLRID, Slug = BasicSlug },
                    Resource = new Resource { LitterRobotId = BasicLRID, UnitStatus = BasicUnitStatus },
                    Expected = new { LRID = BasicLRID, UnitStatus = BasicUnitStatus, Slug = BasicSlug, Found = true }
                },
                new {
                    Q = new SlugMapping { LRID = BasicLRID, Slug = BasicSlug },
                    Resource = new Resource { LitterRobotId = $"{BasicLRID}-fake" , UnitStatus = BasicUnitStatus },
                    Expected = new { LRID = string.Empty, UnitStatus = string.Empty, Slug = string.Empty, Found = false }
                },
            };

            foreach (var test in tests)
            {
                var logger = new Mock<ILogger<MQTTLiason>>();
                var generator = new Mock<IMQTTGenerator>();
                var sharedOpts = Options.Create(new SharedOpts
                {
                    Resources = new[] { test.Q }.ToList(),
                });

                generator.Setup(x => x.BuildDiscovery(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<System.Reflection.AssemblyName>(), false))
                    .Returns(new TwoMQTT.Core.Models.MQTTDiscovery());
                generator.Setup(x => x.StateTopic(test.Q.Slug, nameof(Resource.UnitStatus)))
                    .Returns($"totes/{test.Q.Slug}/topic/{nameof(Resource.UnitStatus)}");

                var mqttLiason = new MQTTLiason(logger.Object, generator.Object, sharedOpts);
                var results = mqttLiason.MapData(test.Resource);
                var actual = results.Where(x => x.topic != null).FirstOrDefault(x => x.topic.Contains(nameof(Resource.UnitStatus)));

                Assert.AreEqual(test.Expected.Found, results.Any(), "The mapping should exist if found.");
                if (test.Expected.Found)
                {
                    Assert.IsTrue(actual.topic.Contains(test.Expected.Slug), "The topic should contain the expected LRID.");
                    Assert.AreEqual(test.Expected.UnitStatus, actual.payload, "The payload be the expected UnitStatus.");
                }
            }
        }

        [TestMethod]
        public void DiscoveriesTest()
        {
            var tests = new[] {
                new {
                    Q = new SlugMapping { LRID = BasicLRID, Slug = BasicSlug },
                    Resource = new Resource { LitterRobotId = BasicLRID, UnitStatus = BasicUnitStatus },
                    Expected = new { LRID = BasicLRID, State = BasicUnitStatus, Slug = BasicSlug }
                },
            };

            foreach (var test in tests)
            {
                var logger = new Mock<ILogger<MQTTLiason>>();
                var generator = new Mock<IMQTTGenerator>();
                var sharedOpts = Options.Create(new SharedOpts
                {
                    Resources = new[] { test.Q }.ToList(),
                });

                generator.Setup(x => x.BuildDiscovery(test.Q.Slug, It.IsAny<string>(), It.IsAny<System.Reflection.AssemblyName>(), false))
                    .Returns(new TwoMQTT.Core.Models.MQTTDiscovery());

                var mqttLiason = new MQTTLiason(logger.Object, generator.Object, sharedOpts);
                var results = mqttLiason.Discoveries();
                var result = results.FirstOrDefault(x => x.sensor == nameof(Resource.UnitStatus));

                Assert.IsNotNull(result, "A discovery should exist.");
            }
        }

        private static string BasicSlug = "totallyaslug";
        private static string BasicUnitStatus = "Rdy";
        private static string BasicLRID = "15873525";
    }
}
